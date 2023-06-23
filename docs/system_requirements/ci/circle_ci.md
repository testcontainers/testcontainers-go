# CircleCI

!!!info
    This document applies to Circle CI Cloud, Server v4.x and Server v3.x.

Your CircleCI configuration should use a dedicated VM for Testcontainers to work. You can achieve this by specifying the 
executor type in your `.circleci/config.yml` to be `machine` instead of the default `docker` executor (see [Choosing an Executor Type](https://circleci.com/docs/executor-intro/) for more info).  

Here is a sample CircleCI configuration that does a checkout of a project and runs Maven:

```yml
jobs:
  build:
    # Check https://circleci.com/docs/executor-intro#linux-vm for more details
    machine: true
      image: ubuntu-2204:2023.04.2
    steps:
      # install Go 1.19
      # checkout the project
      - run: go test./...
```

You can learn more about the best practices of using Testcontainers together with CircleCI in [this article](https://www.atomicjar.com/2022/12/testcontainers-with-circleci/) for Java.
