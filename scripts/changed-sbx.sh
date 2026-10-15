#!/usr/bin/env bash
set -euo pipefail

# ALL_CHANGED_FILES contains both sides of renames from changed-files.
# Use changed-modules.sh for core classification so its exclusions and core
# subpackages stay in sync with the ordinary module and Podman CI detection.
for file in ${ALL_CHANGED_FILES:-}; do
    case "$file" in
        modules/sbx/*|modules/compose/*|\
        .github/workflows/ci.yml|.github/workflows/ci-sbx.yml|\
        .github/workflows/ci-test-go.yml|.github/scripts/*|\
        scripts/changed-sbx.sh|scripts/changed-sbx_test.sh|\
        scripts/test-sbx.sh|.github/scripts/setup-sbx.sh)
            echo true
            exit 0
            ;;
    esac
done

modules=$("$(dirname "${BASH_SOURCE[0]}")/changed-modules.sh")
if jq -e 'index("") != null' <<< "$modules" >/dev/null; then
    echo true
else
    echo false
fi
