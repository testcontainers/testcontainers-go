include ../commons-test.mk

.PHONY: dependencies-scan-examples
dependencies-scan-examples:
	@find . -type f -name Makefile -execdir make dependencies-scan \;

.PHONY: tidy-examples
tidy-examples:
	@find . -type f -name Makefile -execdir make tools-tidy \;
