# Deployment

Topologie tří strojů:

```
┌─────────────┐   HTTPS (CF Access)   ┌──────────────────┐
│     NAS     │ ────────────────────► │    DGX SPARK     │
│  engine     │                       │  LiteLLM         │
│  (Docker)   │                       │  llm-dev (Gemma) │
│  SQLite     │                       │  llm-lab (insights)
│  web :8080  │                       └──────────────────┘
└─────▲───────┘
      │ rsync + docker compose (ssh)
┌─────┴───────┐
│     mac     │  vývoj, build, orchestrace, spouštění simulací přes REST API
└─────────────┘
```

## NAS — web + databáze + Go engine

Engine je jedna statická binárka (pure-Go SQLite, žádné CGO) zabalená v Docker image
(~20 MB, alpine). Databáze i HTML reporty žijí na bind-mount volumes, takže přežijí
rebuild kontejneru.

První nasazení:

```bash
# na NASu (jednorázově)
mkdir -p /volume1/docker/auction-sim
cd /volume1/docker/auction-sim
cp .env.example .env   # a doplň CF_ACCESS_CLIENT_ID / CF_ACCESS_CLIENT_SECRET

# z macu
make deploy-nas NAS_HOST=nas NAS_DIR=/volume1/docker/auction-sim
```

`deploy-nas` synchronizuje zdrojáky přes rsync (bez `.env`, `data/`, `reports/`)
a na NASu spustí `docker compose up -d --build`. `.env` se **nesynchronizuje** —
credentials zůstávají jen na NASu.

Kontejner má healthcheck na `GET /healthz` (ověřuje i dostupnost SQLite).
Po restartu enginu zůstávají hotové simulace dostupné — status i odkaz na HTML
report se čtou z DB, jen JSON detail reportu je pouze pro simulace z aktuálního běhu.

Alternativa bez Dockeru: `make build-linux NAS_ARCH=amd64` (nebo `arm64`) a na NAS
nakopírovat `bin/linux-*/engine` + adresáře `prompts/` a `web/` + `.env`.

## SPARK — AI (LiteLLM, Python)

Engine na Spark nic nedeployuje — jen volá jeho HTTPS endpoint (`LLM_BASE_URL`,
default `https://llm.ol1n.com`) chráněný Cloudflare Access:

- `llm-dev` (Gemma) — rozhodování bidderů, timeout 5 s, fallback WAIT
- `llm-lab` — generování česky psaných insights do reportů, timeout 60 s

Bez platných `CF_ACCESS_*` credentials engine nastartuje, ale zaloguje warning
a všichni bidderi jen čekají (WAIT) → výsledky budou degenerované (no_sale).

Python na Sparku je připravený pro navazující analytiku (plán §16 „po v1"):
export dat ze SQLite (`data/auction.db` na NASu) do CSV/Parquet a stavba
fine-tuning datasetu z uložených (state, decision, reasoning) trojic v tabulce `bids`.

## mac — orchestrace a pipeline

- vývoj + testy: `make build`, `make test`, `make vet`
- cross-compile pro NAS: `make build-linux NAS_ARCH=amd64|arm64` (CGO_ENABLED=0)
- nasazení: `make deploy-nas NAS_HOST=<ssh host> NAS_DIR=<cesta>`
- spouštění simulačních dávek proti NASu přes REST API:

```bash
curl -X POST http://nas:8080/api/v1/simulations \
  -H 'Content-Type: application/json' \
  -d '{"name":"batch-1","item":{"name":"iPhone","retail_price":25000},
       "auction_types":["VICKREY","DUTCH","PENNY"],"runs_per_type":5,
       "bidder_pool":[{"role":"GAMBLER","count":3},{"role":"CAUTIOUS","count":3},
                      {"role":"SNIPER","count":2},{"role":"SOCIAL","count":2}]}'
```

CI (GitHub Actions, `.github/workflows/ci.yml`): build + vet + test + cross-compile
check (amd64/arm64) + docker build na každý push do `main` a každý PR.

## Pozor na case-sensitivity

Prompty se načítají z `prompts/` (malé p). Na macu (case-insensitive APFS) by
prošel i překlep ve velikosti písmen, na NASu (Linux, ext4/btrfs) ne — proto
runtime prompty žijí v `prompts/` a plánovací dokumenty odděleně v `docs/plans/`.
