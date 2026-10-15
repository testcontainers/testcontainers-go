#!/usr/bin/env bash
set -euo pipefail

script="$(dirname "${BASH_SOURCE[0]}")/changed-sbx.sh"
check() {
    local paths=$1 expected=$2 actual
    actual=$(ALL_CHANGED_FILES="$paths" "$script")
    if [[ "$actual" != "$expected" ]]; then
        printf 'Paths: %s\nExpected: %s\nActual: %s\n' "$paths" "$expected" "$actual" >&2
        exit 1
    fi
}

check 'modules/sbx/sbx.go' true
check 'modules/compose/sbx_integration_test.go' true
check 'internal/core/runtime_provider.go' true
check 'docker.go' true
check 'wait/strategy.go' true
check 'network/network.go' true
check 'exec/exec.go' true
check 'log/consumer.go' true
check 'wait/deleted.go' true
# Both sides of a rename are supplied by changed-files.
check 'modules/redis/old.go network/new.go' true
check 'exec/old.go modules/redis/new.go' true
check 'go.mod' true
check 'scripts/test-sbx.sh' true
check '.github/workflows/ci.yml' true
check '.github/workflows/ci-test-go.yml' true
check '.github/scripts/modules/sbx/install-dependencies.sh' true
check '.github/scripts/setup-sbx.sh' true
check '.github/scripts/sbx-test-main.go.tmpl' true
check 'scripts/changed-sbx.sh' true
check 'scripts/changed-sbx_test.sh' true
check 'modules/redis/redis.go' false
check 'modules/podman/podman.go' false
check 'docs/modules/sbx.md README.md mkdocs.yml .devcontainer/config.json' false
check '' false
echo 'sbx change detection tests passed.'
