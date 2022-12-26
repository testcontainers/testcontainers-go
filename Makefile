include ./commons-test.mk

.PHONY: test-all
test-all: tools test-unit

.PHONY: test-examples
test-examples:
	@echo "Running example tests..."
	make -C examples test
