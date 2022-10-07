
.PHONY: test
test: tools
	go run gotest.tools/gotestsum \
		--format short-verbose \
		--rerun-fails=5 \
		--junitfile TEST-e2e.xml \
		--packages="./..." \

.PHONY: tools
tools:
	go mod download

.PHONY: tools-tidy
tools-tidy:
	go mod tidy
