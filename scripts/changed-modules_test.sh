#!/usr/bin/env bash

set -euo pipefail

fixture=$(mktemp -d)
trap 'rm -rf "$fixture"' EXIT
mkdir -p "$fixture"/{scripts,modules/{podman,redis,k6},examples/nginx,modulegen,usage-metrics}
cp "$(dirname "${BASH_SOURCE[0]}")/changed-modules.sh" "$fixture/scripts/"
for module in '' modules/podman modules/redis modules/k6 examples/nginx modulegen usage-metrics; do
    touch "$fixture/$module/go.mod"
done

check() {
    local changed=$1 expected=$2 actual
    actual=$(ALL_CHANGED_FILES="$changed" "$fixture/scripts/changed-modules.sh" 2>/dev/null)
    if [[ "$(jq -c 'sort' <<< "$actual")" != "$(jq -c 'sort' <<< "$expected")" ]]; then
        printf 'Changed: %s\nExpected: %s\nActual: %s\n' "$changed" "$expected" "$actual" >&2
        exit 1
    fi
}

all='["", "examples/nginx", "modulegen", "modules/podman", "modules/redis", "usage-metrics"]'
check 'modules/podman/podman.go' '["modules/podman"]'
check 'modules/podman/go.mod modules/podman/README.md' '["modules/podman"]'
check 'modules/podman/deleted_test.go' '["modules/podman"]'
# A rename supplies both paths, including the old module.
check 'modules/podman/old.go modules/redis/new.go' '["modules/podman", "modules/redis"]'
check 'docker.go' "$all"
check 'deleted_core.go' "$all"
check '.github/workflows/ci-test-go.yml' "$all"
check '.github/scripts/setup-podman.sh' "$all"
check 'docs/modules/podman.md mkdocs.yml .github/dependabot.yml' '[]'
check 'modules/k6/k6.go' '[]'
check '' '[]'
# A removed module must not produce a job whose working directory is missing.
rm "$fixture/modules/podman/go.mod"
check 'modules/podman/go.mod modules/podman/podman.go' '[]'
check 'modules/podman/old.go modules/redis/new.go' '["modules/redis"]'
echo 'Module detection tests passed.'
