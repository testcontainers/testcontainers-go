#!/usr/bin/env bash
set -euo pipefail

# The sbx module's ordinary Docker unit jobs need no VM setup.
if [[ "${TESTCONTAINERS_CI_RUNTIME:-}" != sbx ]]; then
    exit 0
fi

repository_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)
exec "$repository_root/.github/scripts/setup-sbx.sh"
