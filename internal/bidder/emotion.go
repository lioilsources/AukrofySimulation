package bidder

// updateEmotion přepočítá emoci biddera podle zbývajícího rozpočtu a role.
func (b *Bidder) updateEmotion() {
	frac := 1.0
	if b.Budget > 0 {
		frac = b.BudgetRemaining / b.Budget
	}
	switch {
	case b.Won:
		b.Emotion = Satisfied
	case frac < 0.15:
		b.Emotion = Panicked
	case frac < 0.4:
		if b.Persona.RiskTolerance > 0.6 {
			b.Emotion = Determined
		} else {
			b.Emotion = Frustrated
		}
	case b.BidsPlaced > 3 && b.Persona.SunkCostSensitivity > 0.6:
		b.Emotion = Determined
	default:
		b.Emotion = Calm
	}
}
