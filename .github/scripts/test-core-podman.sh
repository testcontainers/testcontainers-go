#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
root_dir=$PWD

# No go.work or optional module import is allowed in these fresh test processes.
export GOWORK=off
make test-tools
for mode in auto explicit; do
    TCGO_PODMAN_COMPAT="$mode" TESTCONTAINERS_PROVIDER='' \
        gotestsum --format short-verbose --packages . \
        --junitfile "TEST-unit-podman-compat-${mode}.xml" -- \
        -count=1 -race -timeout=5m -run '^(TestPodmanCompatibility|TestReaperReapsOrphans)$'
done

# Root ./... stays within the core Go module. A temporary workspace makes the
# optional provider import available without adding it to core's go.mod/go.sum.
workspace_dir=$(mktemp -d)
bootstrap_files=()
cleanup() {
    if ((${#bootstrap_files[@]})); then
        rm -f "${bootstrap_files[@]}"
    fi
    rm -rf "$workspace_dir"
}
trap cleanup EXIT

go list -f '{{.Dir}} {{.Name}}' ./... > "$workspace_dir/packages"
export GOWORK="$workspace_dir/go.work"
go work init "$root_dir" "$root_dir/modules/podman"
while read -r directory package; do
    bootstrap="$directory/zz_podman_ci_test.go"
    test ! -e "$bootstrap"
    bootstrap_files+=("$bootstrap")
    sed "s/^package PACKAGE_test$/package ${package}_test/" \
        .github/scripts/podman-test-main.go.tmpl > "$bootstrap"
done < "$workspace_dir/packages"

# No retries initially: provider/bootstrap failures must remain visible.
make test-unit TEST_RERUN_FAILS=0 GO_TEST_FLAGS=-count=1 \
    TEST_REPORT=TEST-unit-podman.xml TEST_COVERAGE=coverage-podman.out
