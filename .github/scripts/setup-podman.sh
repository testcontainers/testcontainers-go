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

test "$(sudo podman info --format '{{.Host.Security.Rootless}}')" = false

# Runner images can already have an idle socket-activated service. Restart it
# after configuration, and keep it alive across compilation and test processes.
sudo systemctl stop podman.service podman.socket
sudo mkdir -p /etc/systemd/system/podman.service.d /etc/systemd/system/podman.socket.d
sudo tee /etc/systemd/system/podman.service.d/testcontainers.conf > /dev/null <<'SERVICE'
[Service]
ExecStart=
ExecStart=/usr/bin/podman system service --time=0
SERVICE
sudo tee /etc/systemd/system/podman.socket.d/testcontainers.conf > /dev/null <<SOCKET
[Socket]
SocketGroup=$(id -gn)
SocketMode=0660
DirectoryMode=0755
SOCKET
sudo systemctl daemon-reload
sudo systemctl start podman.socket podman.service
sudo chmod 755 /run/podman
curl --fail --silent --show-error --retry 10 --retry-connrefused --retry-delay 1 \
    --retry-max-time 30 --max-time 5 --unix-socket /run/podman/podman.sock http://localhost/version |
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
