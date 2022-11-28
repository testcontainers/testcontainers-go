# Contributing to documentation

The Testcontainers for Go documentation is a static site built with [MkDocs](https://www.mkdocs.org/).
We use the [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) theme, which offers a number of useful extensions to MkDocs.

In addition we use a [custom plugin](https://github.com/rnorth/mkdocs-codeinclude-plugin) for inclusion of code snippets.

We publish our documentation using Netlify.

## Previewing rendered content

### Using Python locally

* Ensure that you have Python 3.8.0 or higher.
* Set up a virtualenv and run `pip install -r requirements.txt` in the `testcontainers-go` root directory.
* Once Python dependencies have been installed, run `mkdocs serve` to start a local auto-updating MkDocs server.

### PR Preview deployments

Note that documentation for pull requests will automatically be published by Netlify as 'deploy previews'.
These deployment previews can be accessed via the `deploy/netlify` check that appears for each pull request.
