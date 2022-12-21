include ./commons-test.mk

.PHONY: test-all
test-all: tools test-unit test-e2e

.PHONY: test-e2e
test-e2e:
	@echo "Running end-to-end tests..."
	make -C e2e test

.PHONY: test-examples
test-examples:
	@echo "Running example tests..."
	make -C examples test
