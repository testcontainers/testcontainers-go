SHELL = /bin/bash

SRC_ROOT := $(shell git rev-parse --show-toplevel)

GOCMD?= go

TOOLS_MOD_DIR    := $(SRC_ROOT)/internal/tools
TOOLS_MOD_REGEX  := "\s+_\s+\".*\""
TOOLS_PKG_NAMES  := $(shell grep -E $(TOOLS_MOD_REGEX) < $(TOOLS_MOD_DIR)/tools.go | tr -d " _\"")
TOOLS_BIN_DIR    := $(SRC_ROOT)/.tools
TOOLS_BIN_NAMES  := $(addprefix $(TOOLS_BIN_DIR)/, $(notdir $(TOOLS_PKG_NAMES)))

$(TOOLS_BIN_DIR):
	mkdir -p $@

$(TOOLS_BIN_NAMES): $(TOOLS_BIN_DIR) $(TOOLS_MOD_DIR)/go.mod
	cd $(TOOLS_MOD_DIR) && GOOS="" GOARCH="" $(GOCMD) build -o $@ -trimpath $(filter %/$(notdir $@),$(TOOLS_PKG_NAMES))

GOLANGCI_LINT       := $(TOOLS_BIN_DIR)/golangci-lint
GOTESTSUM           := $(TOOLS_BIN_DIR)/gotestsum
MOCKERY             := $(TOOLS_BIN_DIR)/mockery

.PHONY: install
install: $(GOLANGCI_LINT) $(GOTESTSUM) $(MOCKERY)

.PHONY: clean
clean:
	rm $(GOLANGCI_LINT)
	rm $(GOTESTSUM)
	rm $(MOCKERY)

.PHONY: dependencies-scan
dependencies-scan:
	@echo ">> Scanning dependencies in $(CURDIR)..."
	go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth --skip-update-check

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --out-format=colored-line-number --path-prefix=. --verbose -c $(SRC_ROOT)/.golangci.yml --fix

.PHONY: generate
generate: $(MOCKERY)
	go generate ./...

.PHONY: test-%
test-%: $(GOTESTSUM)
	@echo "Running $* tests..."
	$(GOTESTSUM) \
		--format short-verbose \
		--rerun-fails=5 \
		--packages="./..." \
		--junitfile TEST-unit.xml \
		-- \
		-v \
		-coverprofile=coverage.out \
		-timeout=30m \
		-race

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: pre-commit
pre-commit: generate tidy lint
