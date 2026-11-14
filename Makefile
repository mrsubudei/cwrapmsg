GOLANGCI_LINT_VERSION = 2.15.0

.PHONY: test
test:
	go test --timeout 10m -race ./...

coverage:
	go test -race -v -coverpkg=./... -coverprofile=profile.out ./...
	go tool cover -func profile.out

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION) run

.PHONY: install
install:
	go install ./cmd/cwrapmsg
	@echo "Installed in $(shell which cwrapmsg)"
