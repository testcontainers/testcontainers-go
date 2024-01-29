include ./commons-test.mk

.PHONY: test-all
test-all: tools test-tools test-unit

.PHONY: test-examples
test-examples:
	@echo "Running example tests..."
	make -C examples test

.PHONY: tidy-all
tidy-all:
	make -C examples tidy-examples
	make -C modules tidy-modules
