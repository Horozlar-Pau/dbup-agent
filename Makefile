VERSION ?= dev
LDFLAGS := -X github.com/Horozlar-Pau/dbup-agent/internal/version.Version=$(VERSION)

.PHONY: build test lint run clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/dbupd ./cmd/dbupd
	go build -ldflags "$(LDFLAGS)" -o bin/dbup ./cmd/dbup

test:
	go test ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1)

run: build
	./bin/dbupd

clean:
	rm -rf bin
