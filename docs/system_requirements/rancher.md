# Using Rancher Desktop

It is possible to use Rancher Desktop to satisfy the system requirements instead of Docker.

**IMPORTANT**: Please ensure you are running an up-to-date version of Rancher Desktop. There were some key fixes made in earlier versions (especially around v1.6). It is highly unlikely you will be able to get Rancher Desktop working with testcontainers if you are on an old version.

The instructions below are written on the assumption that:

1. you wish to run Rancher Desktop without administrative permissions (i.e. without granting `sudo` access a.k.a *"Administrative Access"* setting tickbox in Rancher Desktop is *unticked*).
2. you are running Rancher Desktop on an Apple-silicon device a.k.a M-series processor.

Steps are as follows:

1. In Rancher Desktop change engine from `containerd` to `dockerd (moby)`.
2. In Rancher Desktop set `VZ mode` networking.
3. On macOS CLI (e.g. `Terminal` app), set the following environment variables:

```sh
export DOCKER_HOST=unix://$HOME/.rd/docker.sock
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock
export TESTCONTAINERS_HOST_OVERRIDE=$(rdctl shell ip a show vznat | awk '/inet / {sub("/.*",""); print $2}')
```

As always, remember that environment variables are not persisted unless you add them to the relevant file for your default shell e.g. `~/.zshrc`.

Credit: Thank you to @pdrosos on GitHub.
