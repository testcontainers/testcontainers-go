
.PHONY: test-all
test-all: tools test-unit test-e2e

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	go test ./...

.PHONY: test-e2e
test-e2e:
	@echo "Running end-to-end tests..."
	make -C e2e test

.PHONY: tools
tools:
	go mod download
