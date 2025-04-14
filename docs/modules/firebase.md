# Firebase

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Firebase.

## Adding this module to your project dependencies

Please run the following command to add the Firebase module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/firebase
```

## Usage example

<!--codeinclude-->
[Creating a Firebase container](../../modules/firebase/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Firebase module exposes one entrypoint function to create the Firebase container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*FirebaseContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Firebase container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Firebase Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "ghcr.io/u-health/docker-firebase-emulator:13.29.2")`.

{% include "../features/common_functional_options.md" %}

#### WithRoot

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you need to set a different Firebase root directory, you can use the `WithRoot` option.
E.g. `firebase.Run(context.Background(), "ghcr.io/u-health/docker-firebase-emulator:13.29.2", firebase.WithRoot("testdata/firebase"))`.

!!! warning
    The root directory must be a valid Firebase project directory, including the `firebase.json` file.

#### WithCache

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you need to enable the Firebase cache, you can use the `WithCache` option, which enables the binary cache based on the Testcontainers SessionID (meaningful only when multiple tests are used).

E.g. `firebase.Run(context.Background(), "ghcr.io/u-health/docker-firebase-emulator:13.29.2", firebase.WithCache())`.

### Container Methods

The Firebase container exposes the following methods:
