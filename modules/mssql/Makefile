include ../../commons-test.mk

# Enforce x509negativeserial=1. This is required in case the tests
# use certificates with negative serial numbers.
# Go 1.23+ rejects these certificates without this flag.
# Cf https://pkg.go.dev/crypto/x509#ParseCertificate
export GODEBUG=x509negativeserial=1

.PHONY: test
test:
	$(MAKE) test-mssql
