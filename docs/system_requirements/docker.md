# General Docker requirements

Testcontainers requires a Docker-API compatible container runtime. 
During development, Testcontainers is actively tested against recent versions of Docker on Linux, as well as against Docker Desktop on Mac and Windows. 
These Docker environments are automatically detected and used by Testcontainers without any additional configuration being necessary.

It is possible to configure Testcontainers to work for other Docker setups, such as a remote Docker host or Docker alternatives. 
However, these are not actively tested in the main development workflow, so not all Testcontainers features might be available and additional manual configuration might be necessary. Please see the [Docker host detection](../features/configuration.md#docker-host-detection) section for more information.

If you have further questions about configuration details for your setup or whether it supports running Testcontainers-based tests, 
please contact the Testcontainers team and other users from the Testcontainers community on [Slack](https://slack.testcontainers.org/).
