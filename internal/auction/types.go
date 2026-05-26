package auction

import (
	"context"
	"time"
)

// AuctionType rozlišuje mechanismus aukce.
type AuctionType string

const (
	Dutch   AuctionType = "DUTCH"
	Vickrey AuctionType = "VICKREY"
	Penny   AuctionType = "PENNY"
)

// State je stav životního cyklu aukce.
type State string

const (
	StatePending  State = "PENDING"
	StateRunning  State = "RUNNING"
	StateFinished State = "FINISHED"
)

// Action je rozhodnutí agenta. WAIT = vyčkat (zůstává aktivní), LEAVE = trvale odejít.
type Action string

const (
	ActionBid    Action = "BID"
	ActionWait   Action = "WAIT"
	ActionLeave  Action = "LEAVE"
	ActionBuyNow Action = "BUY_NOW"
)

// Item je dražené zboží.
type Item struct {
	Name        string  `json:"name"`
	RetailPrice float64 `json:"retail_price"`
	Quantity    int     `json:"quantity"`
	Category    string  `json:"category"`
}

// AuctionParams jsou parametry konkrétní aukce (sjednoceno přes typy).
type AuctionParams struct {
	StartPrice       float64       `json:"start_price"`
	ReservePrice     float64       `json:"reserve_price,omitempty"`
	Decrement        float64       `json:"decrement,omitempty"`         // Dutch
	TickInterval     time.Duration `json:"tick_interval,omitempty"`     // Dutch, Penny
	BidFee           float64       `json:"bid_fee,omitempty"`           // Penny
	PriceIncrement   float64       `json:"price_increment,omitempty"`   // Penny
	TimerReset       time.Duration `json:"timer_reset,omitempty"`       // Penny
	SubmissionWindow time.Duration `json:"submission_window,omitempty"` // Vickrey
	PlatformFeePct   float64       `json:"platform_fee_pct"`
	MaxTicks         int           `json:"max_ticks,omitempty"` // hard cap proti nekonečné aukci
}

// ViralMechanism je identifikátor virálního mechanismu (sekce 10.5 plánu).
type ViralMechanism string

const (
	ViralEntryCredit        ViralMechanism = "entry_credit"
	ViralReferralBonus      ViralMechanism = "referral_bonus"
	ViralSocialProof        ViralMechanism = "social_proof"
	ViralCountdownExtension ViralMechanism = "countdown_extension"
	ViralBuyItNow           ViralMechanism = "buy_it_now"
	ViralStreakBonus        ViralMechanism = "streak_bonus"
)

// ViralConfig konfiguruje virální vrstvu nad enginem.
type ViralConfig struct {
	Mechanisms           []ViralMechanism `json:"mechanisms"`
	EntryCreditAmount    float64          `json:"entry_credit_amount,omitempty"`
	ReferralPct          float64          `json:"referral_pct,omitempty"`
	SocialProofThreshold int              `json:"social_proof_threshold,omitempty"`
	CountdownExtendS     time.Duration    `json:"countdown_extend_s,omitempty"`
	CountdownTriggerS    time.Duration    `json:"countdown_trigger_s,omitempty"`
	BuyNowMultiplier     float64          `json:"buy_now_multiplier,omitempty"`
	StreakDiscountPct    float64          `json:"streak_discount_pct,omitempty"`
	StreakThreshold      int              `json:"streak_threshold,omitempty"`
}

// Has vrací true, pokud je mechanismus aktivní.
func (v ViralConfig) Has(m ViralMechanism) bool {
	for _, x := range v.Mechanisms {
		if x == m {
			return true
		}
	}
	return false
}

// Snapshot je obraz stavu aukce předaný agentům k rozhodnutí.
type Snapshot struct {
	AuctionID      string
	Type           AuctionType
	Item           Item
	Params         AuctionParams
	Viral          ViralConfig
	Tick           int
	CurrentPrice   float64
	NextPrice      float64 // Dutch: cena po dalším tiku
	TimerRemaining time.Duration
	LeaderID       string
	LeaderName     string
	ActiveBidders  int
	BidVelocity    float64 // bidů za minutu
	BuyNowEnabled  bool
	BuyNowPrice    float64
}

// Decision je odpověď agenta na Snapshot.
type Decision struct {
	BidderID   string
	Action     Action
	Amount     float64
	Reasoning  string
	Confidence float64
	LatencyMs  int64
}

// BidRecord je zaznamenaný bid pro persistenci a report.
type BidRecord struct {
	BidderID   string
	Amount     float64
	FeePaid    float64
	Action     Action
	Reasoning  string
	Confidence float64
	LatencyMs  int64
	Timestamp  time.Time
}

// Event je doménová událost emitovaná enginem.
type Event struct {
	AuctionID string         `json:"auction_id"`
	Type      string         `json:"type"`
	Tick      int            `json:"tick"`
	Payload   map[string]any `json:"payload"`
	Timestamp time.Time      `json:"timestamp"`
}

// Result je výsledek jednoho běhu aukce.
type Result struct {
	AuctionID  string
	Type       AuctionType
	WinnerID   string
	FinalPrice float64
	BuyNow     bool
	Bids       []BidRecord
	StartedAt  time.Time
	EndedAt    time.Time
}

// Participants je rozhraní k poolu bidderů, které engine používá k rozhodování i ekonomice.
// Implementuje ho internal/bidder.Pool.
type Participants interface {
	ActiveIDs() []string
	Name(id string) string
	// Decide vrátí rozhodnutí pro všechny aktivní bidery (paralelně přes LLM).
	Decide(ctx context.Context, s Snapshot) []Decision
	CanAfford(id string, cost float64) bool
	// Debit strhne útratu (amount = zaplaceno za zboží, fee = nevratný poplatek).
	Debit(id string, amount, fee float64)
	Credit(id string, amount float64) // entry credit, referral
	Leave(id string)
}

// EmitFunc je callback pro emitování událostí.
type EmitFunc func(Event)
