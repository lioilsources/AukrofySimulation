package config

import (
	"os"
	"strconv"
	"time"
)

// Config drží běhové nastavení načtené z prostředí (.env / env vars).
type Config struct {
	EnginePort string
	DBPath     string
	ReportsDir string

	LLMBaseURL        string
	LLMDecisionModel  string
	LLMInsightModel   string
	CFAccessClientID  string
	CFAccessSecret    string
	DecisionTimeout   time.Duration
	InsightTimeout    time.Duration

	TickPacing     string // "live" | "fast"
	RunsPerType    int
	DutchTick      time.Duration
	PennyTimer     time.Duration
	PennyBidFee    float64
	VickreyWindow  time.Duration
	PlatformFeePct float64

	// Viral defaulty
	EntryCredit          float64
	ReferralPct          float64
	SocialProofThreshold int
	CountdownExtend      time.Duration
	CountdownTrigger     time.Duration
	BuyNowMult           float64
	StreakDiscountPct    float64
	StreakThreshold      int
}

// Load sestaví konfiguraci z prostředí s rozumnými defaulty.
func Load() Config {
	return Config{
		EnginePort: env("ENGINE_PORT", "8080"),
		DBPath:     env("DB_PATH", "./data/auction.db"),
		ReportsDir: env("REPORTS_DIR", "./reports"),

		LLMBaseURL:       env("LLM_BASE_URL", "https://llm.ol1n.com"),
		LLMDecisionModel: env("LLM_DECISION_MODEL", "llm-dev"),
		LLMInsightModel:  env("LLM_INSIGHT_MODEL", "llm-lab"),
		CFAccessClientID: env("CF_ACCESS_CLIENT_ID", ""),
		CFAccessSecret:   env("CF_ACCESS_CLIENT_SECRET", ""),
		DecisionTimeout:  msDur("LLM_DECISION_TIMEOUT_MS", 5000),
		InsightTimeout:   msDur("LLM_INSIGHT_TIMEOUT_MS", 60000),

		TickPacing:     env("TICK_PACING", "live"),
		RunsPerType:    envInt("DEFAULT_RUNS_PER_TYPE", 1),
		DutchTick:      msDur("DEFAULT_DUTCH_TICK_MS", 3000),
		PennyTimer:     secDur("DEFAULT_PENNY_TIMER_S", 10),
		PennyBidFee:    envFloat("DEFAULT_PENNY_BID_FEE", 5.0),
		VickreyWindow:  secDur("DEFAULT_VICKREY_WINDOW_S", 60),
		PlatformFeePct: envFloat("DEFAULT_PLATFORM_FEE_PCT", 5.0),

		EntryCredit:          envFloat("DEFAULT_ENTRY_CREDIT", 5.0),
		ReferralPct:          envFloat("DEFAULT_REFERRAL_PCT", 10.0),
		SocialProofThreshold: envInt("DEFAULT_SOCIAL_PROOF_THRESHOLD", 3),
		CountdownExtend:      secDur("DEFAULT_COUNTDOWN_EXT_S", 15),
		CountdownTrigger:     secDur("DEFAULT_COUNTDOWN_TRIGGER_S", 30),
		BuyNowMult:           envFloat("DEFAULT_BUYNOW_MULT", 1.3),
		StreakDiscountPct:    envFloat("DEFAULT_STREAK_DISCOUNT_PCT", 5.0),
		StreakThreshold:      envInt("DEFAULT_STREAK_THRESHOLD", 3),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envFloat(k string, def float64) float64 {
	if v := os.Getenv(k); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func msDur(k string, defMs int) time.Duration {
	return time.Duration(envInt(k, defMs)) * time.Millisecond
}

func secDur(k string, defS int) time.Duration {
	return time.Duration(envInt(k, defS)) * time.Second
}
