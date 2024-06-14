include ../commons-test.mk

.PHONY: dependencies-scan-examples
dependencies-scan-examples:
	@find . -type f -name Makefile -execdir make dependencies-scan \;

.PHONY: lint-examples
lint-examples:
	@find . -type f -name Makefile -execdir make lint \;

.PHONY: tidy-examples
tidy-examples:
	@find . -type f -name Makefile -execdir make tidy \;
