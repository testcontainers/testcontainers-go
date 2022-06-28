
.PHONY: test
test: tools
	go run gotest.tools/gotestsum \
		--format short-verbose \
		--rerun-fails=5 \
		--packages="./..." \


.PHONY: tools
tools:
	go mod download