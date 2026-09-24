# pathly-crossplane

GO ?= go
BIN ?= bin/provider

.PHONY: test build tidy coverage clean

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./... -count=1

coverage:
	$(GO) test ./internal/client/ -coverprofile=covclient.txt -covermode=atomic -count=1
	$(GO) tool cover -func=covclient.txt
	$(GO) test ./... -coverprofile=covall.txt -covermode=atomic -count=1
	$(GO) tool cover -func=covall.txt

build: tidy
	mkdir -p bin
	$(GO) build -o $(BIN) ./cmd/provider

clean:
	rm -rf bin covclient.txt covall.txt coverage.out coverage-client.out
