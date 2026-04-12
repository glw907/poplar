BINARY := poplar

build:
	go build -o $(BINARY) ./cmd/poplar

test:
	go test ./...

vet:
	go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

install:
	GOBIN=$(HOME)/.local/bin go install ./cmd/poplar

check: vet test

clean:
	rm -f $(BINARY)

.PHONY: build test vet lint install check clean
