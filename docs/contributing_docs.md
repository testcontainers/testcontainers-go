# Contributing to documentation

The _Testcontainers for Go_ documentation is a static site built with [MkDocs](https://www.mkdocs.org/).
We use the [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) theme, which offers a number of useful extensions to MkDocs.

We publish our documentation using Netlify.

## Adding code snippets

To include code snippets in the documentation, we use the [codeinclude plugin](https://github.com/rnorth/mkdocs-codeinclude-plugin), which uses the following syntax:

&lt;!--codeinclude--&gt;<br/>
&#91;Human readable title for snippet&#93;(./relative_path_to_example_code.go) targeting_expression<br/>
&#91;Human readable title for snippet&#93;(./relative_path_to_example_code.go) targeting_expression<br/>
&lt;!--/codeinclude--&gt;<br/>

Where each title snippet in the same `codeinclude` block would represent a new tab
in the snippet, and each `targeting_expression` would be:

- `block:someString` or
- `inside_block:someString`

Please refer to the [codeinclude plugin documentation](https://github.com/rnorth/mkdocs-codeinclude-plugin) for more information.

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
