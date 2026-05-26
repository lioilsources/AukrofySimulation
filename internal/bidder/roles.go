package bidder

// DefaultPersona vrací výchozí numerické knoby pro roli.
func DefaultPersona(r Role) Persona {
	switch r {
	case Gambler:
		return Persona{RiskTolerance: 0.9, SocialFactor: 0.6, SunkCostSensitivity: 0.9}
	case Cautious:
		return Persona{RiskTolerance: 0.1, SocialFactor: 0.2, SunkCostSensitivity: 0.1}
	case Interested:
		return Persona{RiskTolerance: 0.5, SocialFactor: 0.4, SunkCostSensitivity: 0.4}
	case Reseller:
		return Persona{RiskTolerance: 0.3, SocialFactor: 0.1, SunkCostSensitivity: 0.05}
	case Collector:
		return Persona{RiskTolerance: 0.6, SocialFactor: 0.3, SunkCostSensitivity: 0.7}
	case Sniper:
		return Persona{RiskTolerance: 0.4, SocialFactor: 0.1, SunkCostSensitivity: 0.2}
	case Social:
		return Persona{RiskTolerance: 0.6, SocialFactor: 0.95, SunkCostSensitivity: 0.5}
	}
	return Persona{RiskTolerance: 0.5, SocialFactor: 0.5, SunkCostSensitivity: 0.5}
}

// DefaultValuationMult vrací násobek retail ceny pro odhad valuace dané role.
func DefaultValuationMult(r Role) float64 {
	switch r {
	case Gambler:
		return 1.2
	case Cautious:
		return 0.9
	case Interested:
		return 1.15
	case Reseller:
		return 0.65
	case Collector:
		return 2.5
	case Sniper:
		return 1.0
	case Social:
		return 1.1
	}
	return 1.0
}
