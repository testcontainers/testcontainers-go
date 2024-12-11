include ../commons-test.mk

.PHONY: dependencies-scan-modules
dependencies-scan-modules:
	@find . -type f -name Makefile -execdir make dependencies-scan \;

.PHONY: lint-modules
lint-modules:
	@find . -type f -name Makefile -execdir make lint \;

.PHONY: tidy-modules
tidy-modules:
	@find . -type f -name Makefile -execdir make tidy \;
