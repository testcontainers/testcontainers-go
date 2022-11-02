.PHONY: test-%
test-%:
	@echo "Running $* tests..."
	go run gotest.tools/gotestsum \
		--format short-verbose \
		--rerun-fails=5 \
		--packages="./..." \
		--junitfile TEST-$*.xml
