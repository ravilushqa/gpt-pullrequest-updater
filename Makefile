.PHONY: init-tools lint test test-coverage precommit help

build:
	go build -o bin/description ./cmd/description && \
	go build -o bin/review      ./cmd/review

# run this once to install tools required for development.
init-tools:
	cd tools && \
	go mod tidy && \
	go mod verify && \
	go generate -x -tags "tools"

# run golangci-lint
lint: init-tools
	./bin/golangci-lint run --timeout=30m ./...

# run go test
test:
	go test -race -count 1 ./...

# run go test with coverage
test-coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# precommit command. run lint, test
precommit: lint test

# show help
help:
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
