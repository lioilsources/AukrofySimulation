// Package bidder spravuje pool LLM agentů, jejich role, emoce a ekonomický stav.
package bidder

// Role určuje chování agenta (system prompt + default persona).
type Role string

const (
	Gambler    Role = "GAMBLER"
	Cautious   Role = "CAUTIOUS"
	Interested Role = "INTERESTED"
	Reseller   Role = "RESELLER"
	Collector  Role = "COLLECTOR"
	Sniper     Role = "SNIPER"
	Social     Role = "SOCIAL"
)

// Emotion je emocionální stav biddera ovlivňující rozhodování.
type Emotion string

const (
	Calm       Emotion = "calm"
	Frustrated Emotion = "frustrated"
	Determined Emotion = "determined"
	Satisfied  Emotion = "satisfied"
	Panicked   Emotion = "panicked"
)

// Persona jsou laditelné numerické knoby předané do promptu (sekce 7 plánu).
type Persona struct {
	RiskTolerance       float64 `json:"risk_tolerance"`
	SocialFactor        float64 `json:"social_factor"`
	SunkCostSensitivity float64 `json:"sunk_cost_sensitivity"`
}

// Bidder je jeden agent s konfigurací a běhovým stavem.
type Bidder struct {
	ID         string
	Name       string
	Role       Role
	Budget     float64
	Valuation  float64
	Persona    Persona
	ReferrerID string // pro referral_bonus

	// běhový stav v rámci aukce
	BudgetRemaining  float64
	Spent            float64
	FeesPaid         float64
	BidsPlaced       int
	Emotion          Emotion
	Active           bool
	Won              bool
	EntryCredit      float64
	ReferralEarnings float64
}

// reset připraví biddera na nový běh aukce.
func (b *Bidder) reset() {
	b.BudgetRemaining = b.Budget
	b.Spent = 0
	b.FeesPaid = 0
	b.BidsPlaced = 0
	b.Emotion = Calm
	b.Active = true
	b.Won = false
	b.EntryCredit = 0
	b.ReferralEarnings = 0
}
