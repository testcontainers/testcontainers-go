.PHONY: dependencies-scan
dependencies-scan:
	@echo ">> Scanning dependencies in $(CURDIR)..."
	go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth --skip-update-check

.PHONY: test-%
test-%:
	@echo "Running $* tests..."
	go test ./... -v

.PHONY: tools
tools:
	go mod download

.PHONY: tools-tidy
tools-tidy:
	go mod tidy
