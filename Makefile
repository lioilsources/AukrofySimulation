.PHONY: build run test vet clean build-linux docker-build docker-up deploy-nas

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

# --- deployment na NAS ---
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

# Nasazení na NAS přes rsync + docker compose (vyžaduje ssh přístup, docker na NASu
# a .env vytvořený na NASu v $(NAS_DIR) podle .env.example — creds se nesynchronizují).
deploy-nas:
	rsync -az --delete \
		--exclude '.git' --exclude 'bin' --exclude 'data' --exclude 'reports' --exclude '.env' \
		./ $(NAS_HOST):$(NAS_DIR)/
	ssh $(NAS_HOST) 'cd $(NAS_DIR) && docker compose up -d --build'
