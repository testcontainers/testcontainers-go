#!/usr/bin/env bash
set -euo pipefail

root_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root_dir"
mode=${1:-all}
case "$mode" in module|core|compose|all) ;; *) echo "usage: $0 {module|core|compose|all}" >&2; exit 2 ;; esac

# Compile-only builds every package with no selected tests, sbx prerequisites,
# tool installation, or real-run XML reports. TestMain cleanup remains lazy.
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
    # Match commons-test.mk's installation convention and make its binary
    # discoverable for local callers whose PATH does not include GOPATH/bin.
    tool_gopath=$(go env GOPATH)
    export GOPATH="${tool_gopath%%:*}"
    export PATH="${tool_gopath%%:*}/bin:$PATH"
    GOWORK=off make test-tools GOPATH="${tool_gopath%%:*}"
fi

export TESTCONTAINERS_SBX_WORKSPACE="$root_dir"
# Keep module metadata immutable; each generated workspace and its sums are
# temporary. Stage subshells isolate GOWORK from other stages and the caller.
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-mod=readonly"
if [[ "$compile_only" == 1 ]]; then
    unset TESTCONTAINERS_SBX_INTEGRATION
else
    export TESTCONTAINERS_SBX_INTEGRATION=1
fi
test_timeout=${TESTCONTAINERS_SBX_GO_TEST_TIMEOUT:-30m}

# Each stage has its own trap so independent stages can continue after failure.
setup_stage() {
    work_dir=$(mktemp -d)
    mkdir -p "$work_dir/bind-root"
    export TESTCONTAINERS_SBX_TEST_BIND_ROOT="$work_dir/bind-root"
    export TESTCONTAINERS_SBX_BIND_ROOTS="${TESTCONTAINERS_SBX_BIND_ROOTS:+$TESTCONTAINERS_SBX_BIND_ROOTS:}$work_dir/bind-root"
    bootstrap_files=()
    trap cleanup EXIT
    trap 'exit 130' INT
    trap 'exit 143' TERM
}
cleanup() {
    if ((${#bootstrap_files[@]})); then
        rm -f -- "${bootstrap_files[@]}"
    fi
    rm -rf -- "$work_dir"
}

run_tests() {
    local dir=$1 report=$2; shift 2
    if [[ "$compile_only" == 1 ]]; then
        echo "Compile-only: $dir (no real tests or XML report)"
        (cd "$dir" && go test -v -count=1 -race -p=1 -timeout="$test_timeout" "$@" -run '^$')
    else
        (cd "$dir" && gotestsum --format short-verbose --junitfile "$report" -- \
            -v -count=1 -race -timeout="$test_timeout" "$@")
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
        return 1
    fi
    for test_file in "$directory"/*_test.go; do
        if [[ -f "$test_file" ]] && grep -Eq '^[[:space:]]*func[[:space:]]+TestMain[[:space:]]*\(' "$test_file"; then
            echo "existing TestMain in $directory; merge sbx cleanup before using this harness" >&2
            return 1
        fi
    done
    # noclobber protects existing files and symlinks, including racing callers.
    if ! (set -o noclobber; : > "$bootstrap"); then
        echo "refusing to overwrite $bootstrap" >&2
        return 1
    fi
    bootstrap_files+=("$bootstrap")
    sed "s/^package PACKAGE_test$/package ${package}_test/" \
        "$root_dir/.github/scripts/sbx-test-main.go.tmpl" > "$bootstrap"
}

run_module() (
    setup_stage
    export TESTCONTAINERS_PROVIDER=sbx GOWORK=off
    run_tests modules/sbx TEST-unit-sbx-module.xml ./...
)

run_core() (
    setup_stage
    export TESTCONTAINERS_PROVIDER=sbx GOWORK=off
    # Capture all root-module packages before adding optional provider imports.
    go list -f '{{.Dir}}{{"\t"}}{{.Name}}' ./... > "$work_dir/core-packages"
    make_workspace core
    while IFS=$'\t' read -r directory package; do
        add_bootstrap "$directory" "$package"
    done < "$work_dir/core-packages"
    if [[ "$compile_only" == 1 ]]; then
        run_tests . TEST-unit-sbx.xml ./...
    else
        # commons-test.mk supplies ./..., -race and -timeout=30m before
        # GO_TEST_FLAGS, so this final timeout preserves local configuration.
        # -p=1 bounds concurrent package sandboxes, leaving test parallelism.
        make test-unit TEST_RERUN_FAILS=0 \
            GO_TEST_FLAGS="-count=1 -p=1 -timeout=$test_timeout" \
            TEST_REPORT=TEST-unit-sbx.xml TEST_COVERAGE=coverage-sbx.out
    fi
)

run_compose_routing() (
    setup_stage
    # This process has no real sbx import; the routing test registers a fake.
    export GOWORK=off TESTCONTAINERS_PROVIDER=docker
    run_tests modules/compose TEST-unit-sbx-compose-routing.xml -run '^TestComposeRuntimeRouting$' ./...
)

run_compose_integration() (
    setup_stage
    make_workspace compose
    export TESTCONTAINERS_PROVIDER=sbx
    add_bootstrap "$root_dir/modules/compose" compose
    run_tests modules/compose TEST-unit-sbx-compose-integration.xml -run '^TestSBXComposeIntegration$' ./...
)

# Calling stages in an if/|| condition disables errexit inside their functions.
# Invoke each in a separate shell context with errexit explicitly restored,
# then collect the status without hiding failures or skipping supplements.
stages=()
case "$mode" in
    module) stages=(module) ;;
    core) stages=(core) ;;
    compose) stages=(compose_routing compose_integration) ;;
    all) stages=(module core compose_routing compose_integration) ;;
esac
status=0
for stage in "${stages[@]}"; do
    echo "=== sbx stage: $stage (compile-only=$compile_only) ==="
    set +e
    (set -e; "run_$stage")
    stage_status=$?
    set -e
    echo "=== sbx stage: $stage exit=$stage_status ==="
    if [[ "$stage_status" != 0 ]]; then
        status=1
    fi
done
exit "$status"
