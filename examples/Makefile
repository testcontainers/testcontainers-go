include ../commons-test.mk

.PHONY: test
test:
	$(MAKE) -C cockroachdb test
	$(MAKE) -C nginx test
	$(MAKE) -C pulsar test
	$(MAKE) -C redis test
