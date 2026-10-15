#!/usr/bin/env bash
set -euo pipefail

root_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root_dir"
mode=${1:-all}
case "$mode" in module|core|compose|all) ;; *) echo "usage: $0 {module|core|compose|all}" >&2; exit 2 ;; esac

# Compile-only validates registration and package wiring without calling sbx.
# All other modes require a real host installation, including the CI job.
compile_only=${TESTCONTAINERS_SBX_COMPILE_ONLY:-0}
if [[ "$compile_only" != 0 && "$compile_only" != 1 ]]; then
    echo 'TESTCONTAINERS_SBX_COMPILE_ONLY must be 0 or 1' >&2
    exit 2
fi
if [[ "$compile_only" == 0 ]]; then
    binary=${TESTCONTAINERS_SBX_BINARY:-sbx}
    command -v "$binary" >/dev/null || { echo "sbx binary not found: $binary" >&2; exit 1; }
    "$binary" version
    if [[ "$(uname -s)" == Linux ]]; then
        test -r /dev/kvm && test -w /dev/kvm || { echo 'sbx requires readable and writable /dev/kvm' >&2; exit 1; }
        grep -qw erofs /proc/filesystems || { echo 'sbx requires the erofs filesystem' >&2; exit 1; }
        command -v mkfs.erofs >/dev/null || { echo 'sbx requires mkfs.erofs' >&2; exit 1; }
    fi
fi

export TESTCONTAINERS_SBX_WORKSPACE="$root_dir"
if [[ "$compile_only" == 1 ]]; then
    unset TESTCONTAINERS_SBX_INTEGRATION
else
    export TESTCONTAINERS_SBX_INTEGRATION=1
fi
test_timeout=${TESTCONTAINERS_SBX_GO_TEST_TIMEOUT:-10m}

work_dir=$(mktemp -d)
bootstrap_files=()
cleanup() {
    if ((${#bootstrap_files[@]})); then
        rm -f -- "${bootstrap_files[@]}"
    fi
    rm -rf -- "$work_dir"
}
trap cleanup EXIT

go_test() {
    local dir=$1; shift
    if [[ "$compile_only" == 1 ]]; then
        # The final -run wins over mode-specific selectors passed by callers.
        (cd "$dir" && go test -count=1 -timeout="$test_timeout" "$@" -run '^$')
    else
        (cd "$dir" && go test -count=1 -timeout="$test_timeout" "$@")
    fi
}

make_workspace() {
    local kind=$1
    mkdir -p "$work_dir/$kind"
    if [[ "$kind" == compose ]]; then
        (cd "$work_dir/$kind" && GOWORK=off go work init "$root_dir" "$root_dir/modules/sbx" "$root_dir/modules/compose")
    else
        (cd "$work_dir/$kind" && GOWORK=off go work init "$root_dir" "$root_dir/modules/sbx")
    fi
    export GOWORK="$work_dir/$kind/go.work"
}

add_bootstrap() {
    local directory=$1 package=$2 bootstrap="$1/zz_sbx_harness_test.go" test_file
    if [[ -e "$bootstrap" || -L "$bootstrap" ]]; then
        echo "refusing to overwrite $bootstrap" >&2
        exit 1
    fi
    for test_file in "$directory"/*_test.go; do
        if [[ -f "$test_file" ]] && grep -q '^func TestMain(' "$test_file"; then
            echo "existing TestMain in $directory; merge sbx cleanup before using this harness" >&2
            exit 1
        fi
    done
    bootstrap_files+=("$bootstrap")
    cat > "$bootstrap" <<EOF
package ${package}_test

import (
    "context"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go/internal/core"
    sbx "github.com/testcontainers/testcontainers-go/modules/sbx"
)

func TestMain(m *testing.M) {
    code := m.Run()
    if !core.HasRegisteredProvider("sbx") {
        fmt.Fprintln(os.Stderr, "sbx provider was not registered")
        code = 1
    }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()
    if err := sbx.Cleanup(ctx); err != nil {
        fmt.Fprintf(os.Stderr, "sbx cleanup: %v\\n", err)
        code = 1
    }
    os.Exit(code)
}
EOF
}

run_module() {
    export TESTCONTAINERS_PROVIDER=sbx
    export GOWORK=off
    go_test modules/sbx ./...
}

run_core() {
    export TESTCONTAINERS_PROVIDER=sbx
    export GOWORK=off
    # Capture package names before adding temporary external-test imports.
    go list -f '{{.Dir}} {{.Name}}' ./... > "$work_dir/core-packages"
    make_workspace core
    while read -r directory package; do
        add_bootstrap "$directory" "$package"
    done < "$work_dir/core-packages"
    # The sbx proxy deliberately rejects host and container network modes, so
    # run named core tests for supported behavior instead of the Docker suite.
    # ./... still compiles and checks provider registration in every package.
    local supported_core_tests='^(TestContainerReturnItsContainerID|TestTwoContainersExposingTheSamePort|TestContainerRespondsWithHttp200ForIndex|TestContainerCreationWithVolumeAndFileWritingToIt|TestExecReturnsNonZeroExitCodeAndOutput|TestDockerContainerCopyToContainer|TestContainerLogsShouldBeWithoutStreamHeader|TestContainerAttachedToNewNetwork|Test_BuildContainerFromDockerfileWithBuildArgs)$'
    go_test . -p 1 ./... -run "$supported_core_tests"
}

run_compose() {
    # Routing tests install a fake provider named sbx. Keep that process
    # separate from the real provider import used for the integration smoke.
    export GOWORK=off TESTCONTAINERS_PROVIDER=docker
    go_test modules/compose -run '^TestComposeRuntimeRouting$' .
    make_workspace compose
    export TESTCONTAINERS_PROVIDER=sbx
    add_bootstrap "$root_dir/modules/compose" compose
    go_test modules/compose -run '^TestSBXComposeIntegration$' .
}

case "$mode" in
    module) run_module ;;
    core) run_core ;;
    compose) run_compose ;;
    all) run_module; run_core; run_compose ;;
esac
