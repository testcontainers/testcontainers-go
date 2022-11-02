include ../commons-test.mk

.PHONY: test
test: tools
	$(MAKE) test-e2e

.PHONY: tools
tools:
	go mod download

.PHONY: tools-tidy
tools-tidy:
	go mod tidy
