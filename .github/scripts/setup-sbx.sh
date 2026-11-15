#!/usr/bin/env bash
set -euo pipefail

# CI-only installation for GitHub-hosted Ubuntu 24.04. The host test script
# deliberately uses the sbx already installed by the developer.
if [[ "$(uname -s)" != Linux || ! -r /etc/os-release ]]; then
    echo 'setup-sbx.sh requires Ubuntu Linux' >&2
    exit 1
fi
. /etc/os-release
if [[ "$ID" != ubuntu || "$VERSION_ID" != 24.04 ]]; then
    echo 'setup-sbx.sh requires Ubuntu 24.04' >&2
    exit 1
fi
if [[ ! -e /dev/kvm ]]; then
    echo 'sbx requires /dev/kvm on this runner' >&2
    exit 1
fi
sudo chmod 666 /dev/kvm
if [[ -e /proc/sys/kernel/apparmor_restrict_unprivileged_userns ]]; then
    sudo sysctl -w kernel.apparmor_restrict_unprivileged_userns=0
fi

sudo apt-get update
sudo apt-get install -y ca-certificates curl erofs-utils e2fsprogs
if ! apt-cache show docker-sbx >/dev/null 2>&1; then
    # GitHub runners may already carry Docker's repository with a different
    # keyring. Reusing it avoids duplicate entries with conflicting Signed-By.
    if sudo grep -Rqs 'download.docker.com/linux/ubuntu' /etc/apt/sources.list /etc/apt/sources.list.d; then
        echo 'Docker apt repository exists but docker-sbx is unavailable' >&2
        exit 1
    fi
    sudo install -m 0755 -d /etc/apt/keyrings
    sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    sudo chmod a+r /etc/apt/keyrings/docker.asc
    echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu ${VERSION_CODENAME} stable" | \
        sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
    sudo apt-get update
fi
sudo apt-get install -y docker-sbx

if ! grep -qw erofs /proc/filesystems; then
    if ! sudo modprobe erofs; then
        sudo apt-get install -y "linux-modules-extra-$(uname -r)"
        sudo modprobe erofs
    fi
fi
command -v mkfs.erofs
command -v mkfs.ext4
test -r /dev/kvm && test -w /dev/kvm
sbx version
