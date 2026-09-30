.PHONY: build run test vet clean build-linux docker-build docker-up \
        ssh-keys ssh-copy-id-nas ssh-copy-id-spark ssh-add \
        deploy-nas deploy-spark spark-status

# --- lokální vývoj (mac) ---

build:
	go build -o bin/engine ./cmd/engine
	go build -o bin/reporter ./cmd/reporter

run: build
	./bin/engine

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf bin/ data/*.db reports/*.html

# --- SSH přístup z macu (klíče + agent) ---
# Dedikované ed25519 klíče pro NAS a SPARK; privátní klíče nikdy neopouští mac.
# Cesty lze přepsat: make deploy-nas SSH_KEY_NAS=~/.ssh/muj_klic
SSH_KEY_NAS   ?= $(HOME)/.ssh/auction-sim_nas
SSH_KEY_SPARK ?= $(HOME)/.ssh/auction-sim_spark

# -i se použije jen pokud klíč existuje — jinak se nechá ssh config / agent
SSH_OPTS_NAS   = $(if $(wildcard $(SSH_KEY_NAS)),-i $(SSH_KEY_NAS),)
SSH_OPTS_SPARK = $(if $(wildcard $(SSH_KEY_SPARK)),-i $(SSH_KEY_SPARK),)

ssh-keys:  ## vygeneruje klíče pro NAS a SPARK, pokud ještě neexistují
	@test -f $(SSH_KEY_NAS)   || ssh-keygen -t ed25519 -f $(SSH_KEY_NAS)   -C "auction-sim mac->nas"
	@test -f $(SSH_KEY_SPARK) || ssh-keygen -t ed25519 -f $(SSH_KEY_SPARK) -C "auction-sim mac->spark"

ssh-copy-id-nas: ssh-keys  ## jednorázově nahraje veřejný klíč na NAS (zeptá se na heslo)
	ssh-copy-id -i $(SSH_KEY_NAS).pub $(NAS_HOST)

ssh-copy-id-spark: ssh-keys  ## jednorázově nahraje veřejný klíč na SPARK
	ssh-copy-id -i $(SSH_KEY_SPARK).pub $(SPARK_HOST)

ssh-add: ssh-keys  ## přidá oba privátní klíče do ssh-agenta (macOS: uloží i do keychain)
	ssh-add --apple-use-keychain $(SSH_KEY_NAS) $(SSH_KEY_SPARK) 2>/dev/null \
		|| ssh-add $(SSH_KEY_NAS) $(SSH_KEY_SPARK)

# --- deployment na NAS (engine: web + SQLite + Go) ---
# Cross-compile z macu: modernc.org/sqlite je pure Go, CGO netřeba.
# Architekturu NASu přepiš přes NAS_ARCH (amd64 pro x86 NAS, arm64 pro ARM).
NAS_ARCH ?= amd64
NAS_HOST ?= nas
NAS_DIR  ?= /volume1/docker/auction-sim

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(NAS_ARCH) go build -trimpath -ldflags="-s -w" -o bin/linux-$(NAS_ARCH)/engine ./cmd/engine
	CGO_ENABLED=0 GOOS=linux GOARCH=$(NAS_ARCH) go build -trimpath -ldflags="-s -w" -o bin/linux-$(NAS_ARCH)/reporter ./cmd/reporter

docker-build:
	docker build -t auction-sim:latest .

docker-up:
	docker compose up -d --build

# Nasazení na NAS přes rsync + docker compose (vyžaduje docker na NASu a .env
# vytvořený na NASu v $(NAS_DIR) podle .env.example — creds se nesynchronizují).
deploy-nas:
	rsync -az --delete -e "ssh $(SSH_OPTS_NAS)" \
		--exclude '.git' --exclude 'bin' --exclude 'data' --exclude 'reports' \
		--exclude '.env' --exclude 'spark' \
		./ $(NAS_HOST):$(NAS_DIR)/
	ssh $(SSH_OPTS_NAS) $(NAS_HOST) 'cd $(NAS_DIR) && docker compose up -d --build'

# --- deployment na SPARK (LiteLLM endpoint) ---
# Synchronizuje adresář spark/ (LiteLLM compose + config) a restartuje službu.
# Reálné API klíče/config drž na Sparku mimo git (viz spark/README.md).
SPARK_HOST ?= spark
SPARK_DIR  ?= /opt/auction-sim/litellm

deploy-spark:
	rsync -az -e "ssh $(SSH_OPTS_SPARK)" \
		--exclude 'config.yaml' \
		spark/ $(SPARK_HOST):$(SPARK_DIR)/
	ssh $(SSH_OPTS_SPARK) $(SPARK_HOST) 'cd $(SPARK_DIR) && docker compose up -d'

spark-status:  ## stav LiteLLM služby na Sparku
	ssh $(SSH_OPTS_SPARK) $(SPARK_HOST) 'cd $(SPARK_DIR) && docker compose ps'
