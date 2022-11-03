include ../commons-test.mk

.PHONY: test
test: test-unit
	$(MAKE) -C cockroachdb test
	$(MAKE) -C nginx test
	$(MAKE) -C pulsar test
	$(MAKE) -C redis test
