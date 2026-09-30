# PLÁN 01 — E2E ověření a nasazení (handoff pro Opus)

> Navazuje na `00-PLAN-auction-sim-claude.md` (směrodatný plán v1) a na review
> implementace ze session „implementation-plan-review". Kód v1 je hotový a
> opravený; tento plán pokrývá ostré ověření, nasazení na NAS + SPARK a
> navazující práce. Pracuj v repu `lioilsources/AukrofySimulation`.

## Výchozí stav (co už je hotové — neopakovat)

- Engine (Go, `cmd/engine`): Vickrey/Dutch/Penny, 7 LLM rolí, virální
  mechanismy, SSE live view, SQLite, HTML reporty s insights. Testy v
  `internal/auction/` zelené.
- Opravený startup bug (prompty v `prompts/`, malé p — na Linuxu case-sensitive),
  persistence GET simulace/reportu přes restart (fallback do DB), `GET /healthz`,
  warning při chybějících CF Access creds, `countdown_trigger_s` v Penny.
- Deployment: `Dockerfile` + `docker-compose.yml` (NAS), `spark/` (LiteLLM
  compose + config šablona), Makefile targety `ssh-keys` / `ssh-copy-id-*` /
  `ssh-add` / `deploy-nas` / `deploy-spark` / `spark-status`, CI workflow.
  Podrobnosti: `docs/DEPLOYMENT.md`, `CLAUDE.md`.

## Prerekvizity (dodá uživatel, bez nich nezačínej fázi 1)

- `.env` s platnými `CF_ACCESS_CLIENT_ID` / `CF_ACCESS_CLIENT_SECRET`
  (service token Cloudflare Access pro `llm.ol1n.com`).
- SSH přístup: `NAS_HOST` (user@host NASu) a `SPARK_HOST` (user@host Sparku).
- Na Sparku reálné backendy modelů (co skutečně servíruje `llm-dev`/`llm-lab`)
  — podle nich upravit `spark/config.example.yaml` → `config.yaml` na Sparku.

## Fáze 1 — první ostrý E2E běh (lokálně, mac)

1. `cp .env.example .env`, doplnit CF creds, `TICK_PACING=fast` pro iteraci.
2. `make run`, přes web (`:8080`) spustit simulaci: všechny 3 typy, výchozí
   bidder pool, 1 běh na typ.
3. Akceptace: aspoň 2 ze 3 aukcí skončí prodejem (ne `no_sale`), v tabulce
   `bids` jsou neprázdné `reasoning` a rozumné `llm_latency_ms` (< 5000),
   report obsahuje česky psané insights (ne fallback text).
4. Diagnostika při problémech (očekávané slabiny):
   - vysoký podíl WAIT → zkontrolovat JSON parsing (`internal/llm/parse.go`)
     proti skutečným odpovědím Gemmy; případně zpřísnit/upravit decision
     prompty v `prompts/decision_*.md`;
   - timeouty → zvednout `LLM_DECISION_TIMEOUT_MS`, snížit počet bidderů;
   - degenerované chování rolí → ladit `prompts/role_*.md` (drž je krátké).
5. Všechny opravy commitovat průběžně s testy tam, kde to jde bez LLM
   (parse na zachycených odpovědích = zlatý fixture test).

## Fáze 2 — nasazení NAS

1. `make ssh-keys && make ssh-copy-id-nas NAS_HOST=… && make ssh-add`.
2. Na NASu vytvořit `$(NAS_DIR)/.env` (creds zůstávají na NASu),
   `make deploy-nas NAS_HOST=… NAS_DIR=…`.
3. Akceptace: `curl http://nas:8080/healthz` → `{"status":"ok"}`; simulace
   spuštěná z webu doběhne; po `docker compose restart` je report simulace
   stále dostupný (persistence fallback).

## Fáze 3 — nasazení SPARK

1. Na Sparku `config.yaml` podle skutečných backendů (viz `spark/README.md`).
2. `make ssh-copy-id-spark SPARK_HOST=…`, `make deploy-spark SPARK_HOST=…`,
   `make spark-status SPARK_HOST=…`.
3. Akceptace: `curl -H "CF-Access-Client-Id: …" -H "CF-Access-Client-Secret: …"
   https://llm.ol1n.com/v1/models` vrací aliasy `llm-dev` a `llm-lab`;
   E2E simulace z NASu proti Sparku projde jako ve fázi 1.

## Fáze 4 — zpevnění (po zeleném E2E)

- Testy mimo `internal/auction/`: parse fixtures (`internal/llm`), metriky a
  agregace (`internal/reporter`), store round-trip (`internal/store`,
  in-memory SQLite), emoce (`internal/bidder`). Cíl: `go test ./...` pokrývá
  každý balík aspoň základně.
- Drobnosti vědomě odložené: JSON detail reportu po restartu vrací jen
  `html_url` (plný JSON jen pro simulace aktuálního běhu) — pokud bude vadit,
  persistovat report JSON do DB při dokončení simulace.

## Fáze 5 — navazující práce (backlog z plánu §16, dle priorit uživatele)

1. Export CSV/Parquet ze SQLite pro Python analytiku na Sparku
   (`cmd/reporter` rozšířit o `-export`).
2. Fine-tuning dataset z trojic (state, decision, reasoning) v tabulce `bids`.
3. Multi-round simulace (stejní bidderi napříč aukcemi, učení), English
   auction jako 4. typ, replay viewer.

## Pravidla práce

- Vycházej z větve `claude/implementation-plan-review-uzhpki`; pokud už je
  mergnutá do `main`, pracuj na nové větvi z `main`.
- Před každým commitem `make test && make vet`; commity malé a česky popsané.
- Nikdy necommituj `.env`, klíče ani reálný `spark/config.yaml`.
