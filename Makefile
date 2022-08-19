
.PHONY: test-all
test-all: tools test-unit test-e2e test-nested-container

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	go run gotest.tools/gotestsum \
		--format short-verbose \
		--rerun-fails=5 \
		--packages="./..." \
		-- -coverprofile=cover.txt

.PHONY: test-nested-container
test-unit-nested:
	docker network inspect testcontainers-custom > /dev/null || docker network create testcontainers-custom
	docker build -t nested-sdk -f testresources/nested.dockerfile .
	docker run --rm -i \
		-v /var/run/docker.sock:/var/run/docker.sock:ro \
		-v `pwd`:/work \
		-w /work \
		--network testcontainers-custom \
		-e CGO_ENABLED=0  \
		nested-sdk go run gotest.tools/gotestsum \
                   		--format short-verbose \
                   		--rerun-fails=5 \
                   		--packages="./..."

.PHONY: test-e2e
test-e2e:
	@echo "Running end-to-end tests..."
	make -C e2e test

.PHONY: tools
tools:
	go mod download

.PHONY: verify
verify:
	go mod verify
	go mod tidy
	go vet ./...