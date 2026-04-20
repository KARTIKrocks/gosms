MODULES = . ./twilio ./sns ./vonage ./msg91
SUB_MODULES = ./twilio ./sns ./vonage ./msg91
MODULE_PATH = github.com/KARTIKrocks/gosms

.PHONY: all test test-race coverage lint lint-fix fmt vet tidy build bench clean ci release-prep release-local

all: tidy fmt vet lint build test

## Build all modules
build:
	@for mod in $(MODULES); do \
		echo "==> Building $$mod"; \
		(cd $$mod && go build ./...) || exit 1; \
	done

## Run tests across all modules
test:
	@for mod in $(MODULES); do \
		echo "==> Testing $$mod"; \
		(cd $$mod && go test ./...) || exit 1; \
	done

## Run tests with race detector
test-race:
	@for mod in $(MODULES); do \
		echo "==> Testing (race) $$mod"; \
		(cd $$mod && go test -race -count=1 ./...) || exit 1; \
	done

## Run tests with coverage and generate report
coverage:
	@go test -race -coverprofile=coverage-core.out -covermode=atomic ./...
	@cd twilio && go test -race -coverprofile=../coverage-twilio.out -covermode=atomic ./...
	@cd sns && go test -race -coverprofile=../coverage-sns.out -covermode=atomic ./...
	@cd vonage && go test -race -coverprofile=../coverage-vonage.out -covermode=atomic ./...
	@cat coverage-core.out > coverage.out
	@tail -n +2 coverage-twilio.out >> coverage.out
	@tail -n +2 coverage-sns.out >> coverage.out
	@tail -n +2 coverage-vonage.out >> coverage.out
	@go tool cover -func=coverage.out | tail -1
	@echo "Full report: go tool cover -html=coverage.out"

## Run linter across all modules
lint:
	@for mod in $(MODULES); do \
		echo "==> Linting $$mod"; \
		(cd $$mod && golangci-lint run --timeout=5m ./...) || exit 1; \
	done

## Run linter with auto-fix across all modules
lint-fix:
	@for mod in $(MODULES); do \
		echo "==> Lint-fixing $$mod"; \
		(cd $$mod && golangci-lint run --fix --timeout=5m ./...) || exit 1; \
	done

## Format code
fmt:
	@gofmt -s -w .
	@goimports -w .

## Run go vet across all modules
vet:
	@for mod in $(MODULES); do \
		echo "==> Vetting $$mod"; \
		(cd $$mod && go vet ./...) || exit 1; \
	done

## Run go mod tidy across all modules
tidy:
	@for mod in $(MODULES); do \
		echo "==> Tidying $$mod"; \
		(cd $$mod && go mod tidy) || exit 1; \
	done

## Run benchmarks
bench:
	@for mod in $(MODULES); do \
		echo "==> Benchmarking $$mod"; \
		(cd $$mod && go test -bench=. -benchmem ./...) || exit 1; \
	done

## Run all CI checks
ci: tidy fmt vet lint test-race

## Remove build artifacts and coverage files
clean:
	@rm -f coverage*.out
	@go clean -cache -testcache

## Prepare sub-modules for release: strip replace directives, set version
## Usage: make release-prep VERSION=v0.1.0
release-prep:
ifndef VERSION
	$(error VERSION is required. Usage: make release-prep VERSION=v0.1.0)
endif
	@for mod in $(SUB_MODULES); do \
		echo "==> release-prep $$mod"; \
		(cd $$mod && \
		go mod edit -dropreplace $(MODULE_PATH) && \
		go mod edit -require $(MODULE_PATH)@$(VERSION)); \
	done
	@echo ""
	@echo "Done! Sub-modules now point to $(MODULE_PATH)@$(VERSION)"
	@echo "Next steps:"
	@echo "  git add -A && git commit -m 'Prepare release $(VERSION)'"
	@echo "  git tag $(VERSION)"
	@echo "  git tag twilio/$(VERSION)"
	@echo "  git tag sns/$(VERSION)"
	@echo "  git tag vonage/$(VERSION)"
	@echo "  git push origin main --tags"

## Restore replace directives for local development after a release
release-local:
	@for mod in $(SUB_MODULES); do \
		echo "==> release-local $$mod"; \
		(cd $$mod && \
		go mod edit -replace $(MODULE_PATH)=../ && \
		go mod tidy); \
	done
	@echo ""
	@echo "Done! Sub-modules restored to local replace directives."
