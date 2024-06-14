# Contributing

* Star the project on [Github](https://github.com/testcontainers/testcontainers-go) and help spread the word :)
* Join our [Slack workspace](http://slack.testcontainers.org)
* [Post an issue](https://github.com/testcontainers/testcontainers-go/issues) if you find any bugs
* Contribute improvements or fixes using a [Pull Request](https://github.com/testcontainers/testcontainers-go/pulls). If you're going to contribute, thank you! Please just be sure to:
    * discuss with the authors on an issue ticket prior to doing anything big.
    * follow the style, naming and structure conventions of the rest of the project.
    * make commits atomic and easy to merge.
    * when updating documentation, please see [our guidance for documentation contributions](contributing_docs.md).
    * when updating the `go.mod` file, please run `make tidy-all` to ensure all modules are updated. It will run `golangci-lint` with the configuration set in the root directory of the project. Please be aware that the lint stage could fail if this is not done.
    * apply format running `make lint`
        * For examples: `make -C examples lint`
        * For modules: `make -C modules lint`
    * verify all tests are passing. Build and test the project with `make test-all` to do this.
        * For a given module or example, go to the module or example directory and run `make test`.
