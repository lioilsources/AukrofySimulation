// Package store persistuje simulace, aukce, bidery, bidy a události do SQLite.
package store

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"time"

	_ "modernc.org/sqlite"

	"github.com/ol1n/auction-sim/internal/auction"
	"github.com/ol1n/auction-sim/internal/bidder"
)

//go:embed schema.sql
var schemaSQL string

type Store struct{ db *sql.DB }

// Open otevře (a migruje) SQLite databázi na dané cestě.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// Simulation je metadata simulace.
type Simulation struct {
	ID           string
	Name         string
	Item         auction.Item
	Viral        auction.ViralConfig
	AuctionTypes string
	RunsPerType  int
	Status       string
	CreatedAt    time.Time
	CompletedAt  *time.Time
}

func (s *Store) CreateSimulation(sim Simulation, poolJSON string) error {
	itemJSON, _ := json.Marshal(sim.Item)
	viralJSON, _ := json.Marshal(sim.Viral)
	_, err := s.db.Exec(
		`INSERT INTO simulations (id,name,item_json,bidder_pool_json,viral_config_json,auction_types,runs_per_type,status,created_at)
		 VALUES (?,?,?,?,?,?,?,?,?)`,
		sim.ID, sim.Name, string(itemJSON), poolJSON, string(viralJSON), sim.AuctionTypes, sim.RunsPerType, sim.Status, sim.CreatedAt,
	)
	return err
}

func (s *Store) UpdateSimulationStatus(id, status string, completed *time.Time) error {
	_, err := s.db.Exec(`UPDATE simulations SET status=?, completed_at=? WHERE id=?`, status, completed, id)
	return err
}

func (s *Store) GetSimulation(id string) (*Simulation, error) {
	row := s.db.QueryRow(`SELECT id,name,item_json,viral_config_json,auction_types,runs_per_type,status,created_at,completed_at FROM simulations WHERE id=?`, id)
	var sim Simulation
	var itemJSON, viralJSON string
	if err := row.Scan(&sim.ID, &sim.Name, &itemJSON, &viralJSON, &sim.AuctionTypes, &sim.RunsPerType, &sim.Status, &sim.CreatedAt, &sim.CompletedAt); err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(itemJSON), &sim.Item)
	json.Unmarshal([]byte(viralJSON), &sim.Viral)
	return &sim, nil
}

// SaveAuction zapíše záznam aukce.
func (s *Store) SaveAuction(simID, auctionID string, typ auction.AuctionType, runIdx int, params auction.AuctionParams, res auction.Result) error {
	paramsJSON, _ := json.Marshal(params)
	_, err := s.db.Exec(
		`INSERT INTO auctions (id,simulation_id,type,run_index,params_json,state,winner_bidder_id,final_price,buy_now,started_at,ended_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		auctionID, simID, string(typ), runIdx, string(paramsJSON), string(auction.StateFinished),
		nullStr(res.WinnerID), res.FinalPrice, res.BuyNow, res.StartedAt, res.EndedAt,
	)
	return err
}

// SaveBids zapíše bidy aukce.
func (s *Store) SaveBids(auctionID string, bids []auction.BidRecord) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO bids (auction_id,bidder_id,amount,fee_paid,action,reasoning,confidence,llm_latency_ms,timestamp) VALUES (?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, b := range bids {
		if _, err := stmt.Exec(auctionID, b.BidderID, b.Amount, b.FeePaid, string(b.Action), b.Reasoning, b.Confidence, b.LatencyMs, b.Timestamp); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// SaveBidderStates zapíše konečné stavy bidderů pro danou aukci.
func (s *Store) SaveBidderStates(auctionID string, bidders []*bidder.Bidder) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO bidder_states
		(bidder_id,bidder_name,role,auction_id,valuation,initial_budget,budget_remaining,spent_total,fees_paid,bids_placed,emotion,won,active,entry_credit_received,referral_earnings,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	now := time.Now()
	for _, b := range bidders {
		if _, err := stmt.Exec(b.ID, b.Name, string(b.Role), auctionID, b.Valuation, b.Budget, b.BudgetRemaining,
			b.Spent, b.FeesPaid, b.BidsPlaced, string(b.Emotion), b.Won, b.Active, b.EntryCredit, b.ReferralEarnings, now); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// SaveEvent zapíše doménovou událost.
func (s *Store) SaveEvent(e auction.Event) error {
	payload, _ := json.Marshal(e.Payload)
	_, err := s.db.Exec(`INSERT INTO events (auction_id,type,payload_json,timestamp) VALUES (?,?,?,?)`,
		e.AuctionID, e.Type, string(payload), e.Timestamp)
	return err
}

// SimRow je řádek přehledu simulace.
type SimRow struct {
	ID        string
	Name      string
	Types     string
	Runs      int
	Status    string
	CreatedAt time.Time
}

// ListSimulations vrátí přehled simulací (nejnovější první).
func (s *Store) ListSimulations() ([]SimRow, error) {
	rows, err := s.db.Query(`SELECT id,name,auction_types,runs_per_type,status,created_at FROM simulations ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SimRow
	for rows.Next() {
		var r SimRow
		if err := rows.Scan(&r.ID, &r.Name, &r.Types, &r.Runs, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
