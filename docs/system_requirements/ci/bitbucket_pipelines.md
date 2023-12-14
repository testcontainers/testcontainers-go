# Bitbucket Pipelines

To enable access to Docker in Bitbucket Pipelines, you need to add `docker` as a service on the step.

Furthermore, Ryuk needs to be turned off since Bitbucket Pipelines does not allow starting privileged containers (see [Disabling Ryuk](../../features/configuration.md#customizing-ryuk-the-resource-reaper)). This can either be done by setting a repository variable in Bitbucket's project settings or by explicitly exporting the variable on a step.

In some cases the memory available to Docker needs to be increased.

Here is a sample Bitbucket Pipeline configuration that does a checkout of a project and runs Go tests:

```yml
image: golang:1.x

pipelines:
  default:
    - step:
        script:
          - export TESTCONTAINERS_RYUK_DISABLED=true
          - go test ./...
        services:
          - docker
definitions:
  services:
    docker:
      memory: 2048
```
