include ../commons-test.mk

.PHONY: dependencies-scan-examples
dependencies-scan-examples:
	@find . -type f -name Makefile -execdir make dependencies-scan \;

.PHONY: test
test: test-unit
	$(MAKE) -C cockroachdb test
	$(MAKE) -C nginx test
	$(MAKE) -C pulsar test
	$(MAKE) -C redis test

.PHONY: tidy-examples
tidy-examples:
	@find . -type f -name Makefile -execdir make tools-tidy \;
