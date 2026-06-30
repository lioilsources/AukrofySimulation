# AukrofySimulation — CLAUDE.md

## Overview

Go simulation of auction-based resource allocation (module: `github.com/ol1n/auction-sim`). Runs agents (bidders) that bid for resources using configurable strategies. Results stored in SQLite, exported as HTML reports. Has a web interface for live monitoring.

## Build & Run

```bash
# Build both binaries
make build          # → bin/engine, bin/reporter

# Run simulation engine
make run            # builds + runs ./bin/engine

# Tests
make test           # go test ./...
make vet            # go vet ./...

# Clean
make clean          # rm -rf bin/ data/*.db reports/*.html
```

## Project Structure

```
cmd/
  engine/           # Simulation engine entry point
  reporter/         # HTML report generator

internal/
  auction/          # Core auction logic (rounds, clearing)
  bidder/           # Agent strategies
  config/           # YAML config loader
  events/           # Event log
  llm/              # Optional LLM-driven bidder strategies
  reporter/         # Report formatting
  store/            # SQLite persistence

data/               # SQLite databases (git-ignored)
reports/            # Generated HTML reports (git-ignored)
web/                # Static frontend templates
  static/
  templates/
```

## Key Design

- Agents are configurable via YAML — strategy type, parameters, initial budget
- Each auction round: agents submit bids → clearing algorithm → allocation
- LLM module: optional, agents can call LLM API for strategic decisions
- SQLite stores full event log for reproducibility and reporting
