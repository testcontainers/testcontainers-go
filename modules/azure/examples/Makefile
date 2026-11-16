include ../../../commons-test.mk

# Preserve the Azure tests' support for certificates with negative serial numbers.
export GODEBUG=x509negativeserial=1

.PHONY: test
test:
	$(MAKE) test-azure-examples
