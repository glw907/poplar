build:
	go build -o beautiful-aerc ./cmd/beautiful-aerc
	go build -o fastmail-cli ./cmd/fastmail-cli
	go build -o tidytext ./cmd/tidytext

test:
	go test ./...

vet:
	go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

install:
	GOBIN=$(HOME)/.local/bin go install ./cmd/beautiful-aerc
	GOBIN=$(HOME)/.local/bin go install ./cmd/fastmail-cli
	GOBIN=$(HOME)/.local/bin go install ./cmd/tidytext

check: vet test

clean:
	rm -f beautiful-aerc fastmail-cli tidytext

.PHONY: build test vet lint install check clean
