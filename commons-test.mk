ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: dependencies-scan
dependencies-scan:
	@echo ">> Scanning dependencies in $(CURDIR)..."
	go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth --skip-update-check

.PHONY: lint
lint:
	golangci-lint run --out-format=github-actions --path-prefix=. --verbose -c $(ROOT_DIR)/.golangci.yml --fix

.PHONY: test-%
test-%:
	@echo "Running $* tests..."
	gotestsum \
		--format short-verbose \
		--rerun-fails=5 \
		--packages="./..." \
		--junitfile TEST-unit.xml \
		-- \
		-coverprofile=coverage.out \
		-timeout=30m

.PHONY: tools
tools:
	go mod download

.PHONY: tools-tidy
tools-tidy:
	go mod tidy
