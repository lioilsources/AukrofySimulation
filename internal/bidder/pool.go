package bidder

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ol1n/auction-sim/internal/auction"
	"github.com/ol1n/auction-sim/internal/llm"
)

// Pool spravuje bidder agenty a implementuje auction.Participants.
type Pool struct {
	bidders map[string]*Bidder
	order   []string

	client  *llm.Client
	reg     *llm.Registry
	model   string
	temp    float64
	timeout time.Duration
}

// NewPool vytvoří pool z bidderů.
func NewPool(bidders []*Bidder, client *llm.Client, reg *llm.Registry, model string, temp float64, timeout time.Duration) *Pool {
	p := &Pool{
		bidders: map[string]*Bidder{},
		client:  client,
		reg:     reg,
		model:   model,
		temp:    temp,
		timeout: timeout,
	}
	for _, b := range bidders {
		p.bidders[b.ID] = b
		p.order = append(p.order, b.ID)
	}
	return p
}

// ResetForAuction připraví všechny bidery na nový běh.
func (p *Pool) ResetForAuction() {
	for _, b := range p.bidders {
		b.reset()
	}
}

// Bidders vrací bidery v deterministickém pořadí (pro report).
func (p *Pool) Bidders() []*Bidder {
	out := make([]*Bidder, 0, len(p.order))
	for _, id := range p.order {
		out = append(out, p.bidders[id])
	}
	return out
}

func (p *Pool) Get(id string) *Bidder { return p.bidders[id] }

// --- auction.Participants ---

func (p *Pool) ActiveIDs() []string {
	var ids []string
	for _, id := range p.order {
		if p.bidders[id].Active {
			ids = append(ids, id)
		}
	}
	return ids
}

func (p *Pool) Name(id string) string {
	if b := p.bidders[id]; b != nil {
		return b.Name
	}
	return id
}

func (p *Pool) CanAfford(id string, cost float64) bool {
	b := p.bidders[id]
	return b != nil && b.BudgetRemaining >= cost
}

func (p *Pool) Debit(id string, amount, fee float64) {
	b := p.bidders[id]
	if b == nil {
		return
	}
	b.BudgetRemaining -= amount + fee
	b.Spent += amount
	b.FeesPaid += fee
	if amount > 0 || fee > 0 {
		b.BidsPlaced++
	}
	if amount > 0 {
		b.Won = true
	}
	b.updateEmotion()
}

func (p *Pool) Credit(id string, amount float64) {
	b := p.bidders[id]
	if b == nil {
		return
	}
	b.BudgetRemaining += amount
	b.EntryCredit += amount
}

func (p *Pool) Leave(id string) {
	if b := p.bidders[id]; b != nil {
		b.Active = false
		b.updateEmotion()
	}
}

// Decide získá rozhodnutí všech aktivních bidderů paralelně přes LLM.
func (p *Pool) Decide(ctx context.Context, s auction.Snapshot) []auction.Decision {
	ids := p.ActiveIDs()
	decisions := make([]auction.Decision, len(ids))

	dctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	g, gctx := errgroup.WithContext(dctx)
	for i, id := range ids {
		i, b := i, p.bidders[id]
		g.Go(func() error {
			d := p.decideOne(gctx, b, s)
			d.BidderID = b.ID
			decisions[i] = d
			return nil
		})
	}
	_ = g.Wait()
	return decisions
}

func (p *Pool) decideOne(ctx context.Context, b *Bidder, s auction.Snapshot) auction.Decision {
	system := p.systemPrompt(b)
	data := p.promptData(b, s)
	user, err := p.reg.RenderDecision(s.Type, data)
	if err != nil {
		return auction.Decision{Action: auction.ActionWait, Reasoning: "chyba šablony", Confidence: 0.5}
	}
	raw, latency, err := p.client.Complete(ctx, p.model, system, user, 256, p.temp)
	if err != nil {
		return auction.Decision{Action: auction.ActionWait, Reasoning: "LLM timeout/error", Confidence: 0.5, LatencyMs: latency}
	}
	d := llm.ParseDecision(raw)
	d.LatencyMs = latency
	return d
}

func (p *Pool) systemPrompt(b *Bidder) string {
	base := p.reg.RoleSystem(string(b.Role))
	return fmt.Sprintf("%s\n\nTVÉ PARAMETRY:\n- Budget: %.0f Kč\n- Valuace zboží: %.0f Kč\n- risk_tolerance: %.2f (0=opatrný, 1=hazardér)\n- social_factor: %.2f (0=samotář, 1=virální)\n- sunk_cost_sensitivity: %.2f (náchylnost k utopeným nákladům)",
		base, b.Budget, b.Valuation, b.Persona.RiskTolerance, b.Persona.SocialFactor, b.Persona.SunkCostSensitivity)
}

// PromptData je sjednocený kontext předaný decision šablonám.
type PromptData struct {
	Item                 auction.Item
	CurrentPrice         float64
	NextPrice            float64
	BidFee               float64
	PriceIncrement       float64
	TimerRemaining       int
	TickRemaining        int
	LeaderName           string
	ActiveBidders        int
	BidVelocity          float64
	BudgetRemaining      float64
	Valuation            float64
	YourBids             int
	YourFees             float64
	Emotion              string
	RecentHistory        string
	EntryCreditAvailable float64
	BuyNowEnabled        bool
	BuyNowPrice          float64
}

func (p *Pool) promptData(b *Bidder, s auction.Snapshot) PromptData {
	leader := s.LeaderName
	if leader == "" {
		leader = "nikdo"
	}
	hist := fmt.Sprintf("leader: %s, cena: %.2f Kč, aktivních: %d", leader, s.CurrentPrice, s.ActiveBidders)
	return PromptData{
		Item:                 s.Item,
		CurrentPrice:         s.CurrentPrice,
		NextPrice:            s.NextPrice,
		BidFee:               s.Params.BidFee,
		PriceIncrement:       s.Params.PriceIncrement,
		TimerRemaining:       int(s.TimerRemaining.Seconds()),
		TickRemaining:        int(s.Params.TickInterval.Seconds()),
		LeaderName:           leader,
		ActiveBidders:        s.ActiveBidders,
		BidVelocity:          s.BidVelocity,
		BudgetRemaining:      b.BudgetRemaining,
		Valuation:            b.Valuation,
		YourBids:             b.BidsPlaced,
		YourFees:             b.FeesPaid,
		Emotion:              string(b.Emotion),
		RecentHistory:        hist,
		EntryCreditAvailable: b.EntryCredit,
		BuyNowEnabled:        s.BuyNowEnabled,
		BuyNowPrice:          s.BuyNowPrice,
	}
}
