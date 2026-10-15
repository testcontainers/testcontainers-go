#!/usr/bin/env bash

# This runs only on the disposable Ubuntu runner used by the core Podman job.
set -euo pipefail

sudo apt-get update -qq
sudo apt-get install -y podman
podman --version

# Do not allow endpoint discovery to silently fall back to the runner's Docker.
sudo systemctl stop docker.service docker.socket
sudo systemctl mask --runtime docker.service docker.socket

# Tests use Docker Hub short names. Never prompt for a registry on CI.
sudo mkdir -p /etc/containers/registries.conf.d
sudo tee /etc/containers/registries.conf.d/99-testcontainers.conf > /dev/null <<'REGISTRIES'
unqualified-search-registries = ["docker.io"]
short-name-mode = "disabled"

# The authentication tests start a local HTTP registry on a random port.
[[registry]]
prefix = "localhost"
location = "localhost"
insecure = true
REGISTRIES

sudo systemctl start podman.socket
sudo chgrp "$(id -gn)" /run/podman/podman.sock
sudo chmod 660 /run/podman/podman.sock

test "$(sudo podman info --format '{{.Host.Security.Rootless}}')" = false
curl --fail --silent --show-error --unix-socket /run/podman/podman.sock http://localhost/version |
    jq -e '.Components | any(.Name == "Podman Engine")'

cat >> "$GITHUB_ENV" <<'ENV'
DOCKER_HOST=unix:///run/podman/podman.sock
DOCKER_CONTEXT=
DOCKER_TLS_VERIFY=
DOCKER_CERT_PATH=
XDG_RUNTIME_DIR=
TESTCONTAINERS_PROVIDER=podman
TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/run/podman/podman.sock
TESTCONTAINERS_RYUK_DISABLED=false
TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED=true
ENV
