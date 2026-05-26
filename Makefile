.PHONY: build run test vet clean

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
