#!/usr/bin/env bash
set -euo pipefail

OUTPUT_PATH=""
MODE=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --output_path)
            OUTPUT_PATH="$2"
            shift 2
            ;;
        base|new)
            MODE="$1"
            shift
            ;;
        *)
            shift
            ;;
    esac
done

if [[ -z "$OUTPUT_PATH" ]]; then
    echo "Error: --output_path is required" >&2
    exit 1
fi

if [[ -z "$MODE" ]]; then
    echo "Error: mode (base or new) is required" >&2
    exit 1
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT/modules/postgres"

case "$MODE" in
    base)
        # Regression check: existing tests covering the postgres module's wait-strategy
        # path and basic container startup. Excludes the new reuse test so this mode
        # is independent of the solution patch.
        go test -v -count=1 -timeout 10m \
            -run "^(TestContainerWithWaitForSQL|TestWithConfigFile|TestWithInitScript|TestWithOrderedInitScript)$" \
            ./... 2>&1 \
            | go-junit-report -set-exit-code > "$OUTPUT_PATH"
        ;;
    new)
        # Regression test for false-positive ready signal on reused containers.
        # Fails on the base commit (log wait satisfied by stale logs before crash
        # recovery completes); passes once BasicWaitStrategies adds a live-state probe.
        go test -v -count=1 -timeout 10m \
            -run "^TestBasicWaitStrategies_reusedContainer$" \
            ./... 2>&1 \
            | go-junit-report -set-exit-code > "$OUTPUT_PATH"
        ;;
    *)
        echo "Error: mode must be 'base' or 'new'" >&2
        exit 1
        ;;
esac
