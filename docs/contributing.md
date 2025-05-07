# Contributing

`Testcontainers for Go` is open source, and we love to receive contributions from our community — you!

There are many ways to contribute, from writing tutorials or blog posts, improving the documentation, submitting bug reports and feature requests, or writing code for the core library or for a technology module.

In any case, if you like the project, please star the project on [GitHub](https://github.com/testcontainers/testcontainers-go/stargazers) and help spread the word :)
Also join our [Slack workspace](http://slack.testcontainers.org) to get help, share your ideas, and chat with the community.

## Questions

GitHub is reserved for bug reports and feature requests; it is not the place for general questions.
If you have a question or an unconfirmed bug, please visit our [Slack workspace](https://testcontainers.slack.com/);
feedback and ideas are always welcome.

## Code contributions

If you have a bug fix or new feature that you would like to contribute, please find or open an [issue](https://github.com/testcontainers/testcontainers-go/issues) first.
It's important to talk about what you would like to do, as there may already be someone working on it,
or there may be context to be aware of before implementing the change.

Next would be to fork the repository and make your changes in a feature branch. **Please do not commit changes to the `main` branch**,
otherwise we won't be able to contribute to your changes directly in the PR.

### Submitting your changes

Please just be sure to:

* follow the style, naming and structure conventions of the rest of the project.
* make commits atomic and easy to merge.
* use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) for the PR title. This will help us to understand the nature of the changes, and to generate the changelog after all the commits in the PR are squashed.
    * Please use the `feat!`, `chore!`, `fix!`... types for breaking changes, as these categories are considered as `breaking change` in the changelog. Please use the `!` to denote a breaking change.
    * Please use the `security` type for security fixes, as these categories are considered as `security` in the changelog.
    * Please use the `feat` type for new features, as these categories are considered as `feature` in the changelog.
    * Please use the `fix` type for bug fixes, as these categories are considered as `bug` in the changelog.
    * Please use the `docs` type for documentation updates, as these categories are considered as `documentation` in the changelog.
    * Please use the `chore` type for housekeeping commits, including `build`, `ci`, `style`, `refactor`, `test`, `perf` and so on, as these categories are considered as `chore` in the changelog.
    * Please use the `deps` type for dependency updates, as these categories are considered as `dependencies` in the changelog.

!!!important
    There is a GitHub Actions workflow that will check if your PR title follows the conventional commits convention. If not, it contributes a failed check to your PR.
    To know more about the conventions, please refer to the [workflow file](https://github.com/testcontainers/testcontainers-go/blob/main/.github/workflows/conventions.yml).

* use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) for your commit messages, as it improves the readability of the commit history, and the review process. Please follow the above conventions for the PR title.
* unless necessary, please try to **avoid pushing --force** to the published branch you submitted a PR from, as it makes it harder to review the changes from a given previous state.
* apply format running `make lint-all`. It will run `golangci-lint` for the core and modules with the configuration set in the root directory of the project. Please be aware that the lint stage on CI could fail if this is not done.
    * For linting just the modules: `make -C modules lint-modules`
    * For linting just the examples: `make -C examples lint-examples`
    * For linting just the modulegen: `make -C modulegen lint`
* verify all tests are passing. Build and test the project with `make test-all` to do this.
    * For a given module or example, go to the module or example directory and run `make test`.
    * If you find an `ld warning` message on MacOS, you can ignore it. It is indeed a warning: https://github.com/golang/go/issues/61229
> === Errors
> ld: warning: '/private/var/folders/3y/8hbf585d4yl6f8j5yzqx6wz80000gn/T/go-link-2319589277/000018.o' has malformed LC_DYSYMTAB, expected 98 undefined symbols to start at index 1626, found 95 undefined symbols starting at index 1626

* when updating the `go.mod` file, please run `make tidy-all` to ensure all modules are updated.

## Documentation contributions

The _Testcontainers for Go_ documentation is a static site built with [MkDocs](https://www.mkdocs.org/).
We use the [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) theme, which offers a number of useful extensions to MkDocs.

We publish our documentation using Netlify.

### Adding code snippets

To include code snippets in the documentation, we use the [codeinclude plugin](https://github.com/rnorth/mkdocs-codeinclude-plugin), which uses the following syntax:

> &lt;!--codeinclude--&gt;<br/>
> &#91;Human readable title for snippet&#93;(./relative_path_to_example_code.go) targeting_expression<br/>
> &#91;Human readable title for snippet&#93;(./relative_path_to_example_code.go) targeting_expression<br/>
> &lt;!--/codeinclude--&gt;<br/>

Where each title snippet in the same `codeinclude` block would represent a new tab
in the snippet, and each `targeting_expression` would be:

- `block:someString` or
- `inside_block:someString`

Please refer to the [codeinclude plugin documentation](https://github.com/rnorth/mkdocs-codeinclude-plugin) for more information.

### Previewing rendered content

From the root directory of the repository, you can use the following command to build and serve the documentation locally:

```shell
make serve-docs
```

It will use a Docker container to install the required dependencies and start a local server at `http://localhost:8000`.

Once finished, you can destroy the container with the following command:

```shell
make clean-docs
```

### PR Preview deployments

Note that documentation for pull requests will automatically be published by Netlify as 'deploy previews'.
These deployment previews can be accessed via the `deploy/netlify` check that appears for each pull request.

Please check the GitHub comment Netlify posts on the PR for the URL to the deployment preview.
