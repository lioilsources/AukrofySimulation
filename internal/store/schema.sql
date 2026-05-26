CREATE TABLE IF NOT EXISTS simulations (
    id                TEXT PRIMARY KEY,
    name              TEXT NOT NULL,
    item_json         TEXT NOT NULL,
    bidder_pool_json  TEXT NOT NULL,
    viral_config_json TEXT,
    auction_types     TEXT NOT NULL,
    runs_per_type     INTEGER NOT NULL DEFAULT 1,
    status            TEXT NOT NULL,
    created_at        DATETIME NOT NULL,
    completed_at      DATETIME
);

CREATE TABLE IF NOT EXISTS auctions (
    id               TEXT PRIMARY KEY,
    simulation_id    TEXT NOT NULL REFERENCES simulations(id),
    type             TEXT NOT NULL,
    run_index        INTEGER NOT NULL,
    params_json      TEXT NOT NULL,
    state            TEXT NOT NULL,
    winner_bidder_id TEXT,
    final_price      REAL,
    buy_now          BOOLEAN NOT NULL DEFAULT FALSE,
    started_at       DATETIME,
    ended_at         DATETIME
);

CREATE TABLE IF NOT EXISTS bidder_states (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    bidder_id             TEXT NOT NULL,
    bidder_name           TEXT NOT NULL,
    role                  TEXT NOT NULL,
    auction_id            TEXT NOT NULL REFERENCES auctions(id),
    valuation             REAL NOT NULL,
    initial_budget        REAL NOT NULL,
    budget_remaining      REAL NOT NULL,
    spent_total           REAL NOT NULL,
    fees_paid             REAL NOT NULL DEFAULT 0,
    bids_placed           INTEGER NOT NULL DEFAULT 0,
    emotion               TEXT NOT NULL,
    won                   BOOLEAN NOT NULL DEFAULT FALSE,
    active                BOOLEAN NOT NULL DEFAULT TRUE,
    entry_credit_received REAL NOT NULL DEFAULT 0,
    referral_earnings     REAL NOT NULL DEFAULT 0,
    updated_at            DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS bids (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    auction_id     TEXT NOT NULL REFERENCES auctions(id),
    bidder_id      TEXT NOT NULL,
    amount         REAL NOT NULL,
    fee_paid       REAL NOT NULL DEFAULT 0,
    action         TEXT NOT NULL,
    reasoning      TEXT,
    confidence     REAL,
    llm_latency_ms INTEGER,
    timestamp      DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    auction_id   TEXT NOT NULL,
    type         TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    timestamp    DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bids_auction ON bids(auction_id);
CREATE INDEX IF NOT EXISTS idx_events_auction ON events(auction_id);
CREATE INDEX IF NOT EXISTS idx_states_auction ON bidder_states(auction_id);
CREATE INDEX IF NOT EXISTS idx_auctions_sim ON auctions(simulation_id);
