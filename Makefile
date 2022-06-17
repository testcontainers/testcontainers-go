
.PHONY: test
test: tools
	go test ./...
	echo "Running end-to-end tests..." && cd ./e2e && make

.PHONY: tools
tools:
	go mod download
