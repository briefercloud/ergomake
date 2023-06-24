.PHONY: mocks
mocks:
	rm -rf mocks && mockery --all --keeptree --with-expecter --dir internal

.PHONY: fmt
fmt:
	goimports -local=github.com/ergomake/ergomake -w cmd internal e2e

.PHONY: lint
lint:
	go vet ./...
	staticcheck ./...

.PHONY: tidy
tidy:
	go mod tidy

TESTS = ./internal/...
.PHONY: test
test:
	go test -v -race $(TESTS)

.PHONY: e2e-test
e2e-test:
	go test -v -race -timeout=15m ./e2e/...

COVERPKG = ./internal/...
.PHOHY: coverage
coverage:
	go test -v -race -covermode=atomic -coverprofile cover.out -coverpkg $(COVERPKG) $(TESTS)

.PHONY: deps
deps:
	go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps
	go install golang.org/x/tools/cmd/goimports@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/vektra/mockery/v2@v2.26.1

.PHONY: migrate
migrate:
	go run cmd/migrator/main.go
