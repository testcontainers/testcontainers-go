# CircleCI

!!!info
    This document applies to Circle CI Cloud, Server v4.x and Server v3.x.

Your CircleCI configuration should use a dedicated VM for Testcontainers to work. You can achieve this by specifying the 
executor type in your `.circleci/config.yml` to be `machine` instead of the default `docker` executor (see [Choosing an Executor Type](https://circleci.com/docs/executor-intro/) for more info).  

Here is a sample CircleCI configuration that does a checkout of a project and runs `go test` for a project. Go is installed for the `tests` job using [`gvm`](https://github.com/andrewkroh/gvm), and a workflow matrix has been defined to run the job with different Go versions. Go steps are finally executed from the `go` orb.

```yml
version: 2.1

orbs:
  go: circleci/go@1.11.0

executors:
 machine_executor_amd64:
   machine:
     image: ubuntu-2204:2024.01.2
   environment:
     architecture: "amd64"
     platform: "linux/amd64"

jobs:
  tests:
    executor: machine_executor_amd64
    parameters:
      go-version:
        type: string
    steps:
      - run:
          name: Install GVM
          command: |
            mkdir ~/gvmbin
            curl -sL -o ~/gvmbin/gvm https://github.com/andrewkroh/gvm/releases/download/v0.5.2/gvm-linux-amd64
            chmod +x ~/gvmbin/gvm
            echo 'export PATH=$PATH:~/gvmbin' >> "$BASH_ENV"
      - run:
          name: Install Go
          command: |
            eval "$(gvm << parameters.go-version >>)"
            echo 'eval "$(gvm << parameters.go-version >>)"' >> "$BASH_ENV"
            go version
      - checkout # checkout source code
      - go/load-cache # Load cached Go modules.
      - go/mod-download # Run 'go mod download'.
      - go/save-cache # Save Go modules to cache.
      - go/test: # Runs 'go test ./...' but includes extensive parameterization for finer tuning.
          covermode: atomic
          failfast: true
          race: true

workflows:
  build-and-test:
    jobs:
      - tests:
          matrix:
            parameters:
              go-version: ["1.23.6", "1.24.0"]

```

You can learn more about the best practices of using Testcontainers together with CircleCI in [this article](https://www.atomicjar.com/2022/12/testcontainers-with-circleci/) for Java.
