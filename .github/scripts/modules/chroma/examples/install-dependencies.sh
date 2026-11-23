#!/usr/bin/env bash
set -euo pipefail

runtime_dir=$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/chroma-onnxruntime.XXXXXX")

# PyPI's wheel also ships the native library, bypassing the broken GitHub download.
# Only the native C library is needed. Select a supported wheel even when the
# runner's Python version is newer than ONNX Runtime 1.22.0 supports.
python3 -m pip download --disable-pip-version-check --only-binary=:all: --no-deps \
    --python-version 3.12 --dest "$runtime_dir" onnxruntime==1.22.0 >&2

library_path=$(python3 - "$runtime_dir" <<'PY'
import pathlib
import sys
import zipfile

directory = pathlib.Path(sys.argv[1])
wheel, = directory.glob("onnxruntime-1.22.0-*.whl")
libraries = []
with zipfile.ZipFile(wheel) as archive:
    for member in archive.namelist():
        path = pathlib.PurePosixPath(member)
        if str(path.parent) != "onnxruntime/capi":
            continue
        if path.name.startswith("libonnxruntime") or path.name == "onnxruntime.dll":
            target = directory / path.name
            target.write_bytes(archive.read(member))
            if "providers" not in path.name:
                libraries.append(target)
library, = libraries
print(library.resolve())
PY
)

if [[ -n "${GITHUB_ENV:-}" ]]; then
    printf 'CHROMAGO_ONNX_RUNTIME_PATH=%s\n' "$library_path" >> "$GITHUB_ENV"
else
    printf 'export CHROMAGO_ONNX_RUNTIME_PATH=%q\n' "$library_path"
fi
