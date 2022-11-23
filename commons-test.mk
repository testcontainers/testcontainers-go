.PHONY: test-%
test-%:
	@echo "Running $* tests..."
	go run gotest.tools/gotestsum \
		--format short-verbose \
		--rerun-fails=5 \
		--packages="./..." \
		--junitfile TEST-$*.xml

.PHONY: tools
tools:
	go mod download

.PHONY: tools-tidy
tools-tidy:
	go mod tidy
