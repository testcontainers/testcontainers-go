include ../commons-test.mk

.PHONY: dependencies-scan-modules
dependencies-scan-modules:
	@find . -type f -name Makefile -execdir make dependencies-scan \;

.PHONY: tidy-modules
tidy-modules:
	@find . -type f -name Makefile -execdir make tools-tidy \;
