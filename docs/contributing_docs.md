# Contributing to documentation

The _Testcontainers for Go_ documentation is a static site built with [MkDocs](https://www.mkdocs.org/).
We use the [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) theme, which offers a number of useful extensions to MkDocs.

In addition we use a [custom plugin](https://github.com/rnorth/mkdocs-codeinclude-plugin) for inclusion of code snippets.

We publish our documentation using Netlify.

## Previewing rendered content

### Using Python locally

From the root directory of the repository, you can use the following command to build and serve the documentation locally:

```shell
make serve-docs
```

It will use a Python's virtual environment to install the required dependencies and start a local server at `http://localhost:8000`.

Once finished, you can destroy the virtual environment with the following command:

```shell
make clean-docs
```

### PR Preview deployments

Note that documentation for pull requests will automatically be published by Netlify as 'deploy previews'.
These deployment previews can be accessed via the `deploy/netlify` check that appears for each pull request.
