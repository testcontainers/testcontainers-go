# Contributing

* Star the project on [Github](https://github.com/testcontainers/testcontainers-go) and help spread the word :)
* Join our [Slack group](http://slack.testcontainers.org)
* [Post an issue](https://github.com/testcontainers/testcontainers-go/issues) if you find any bugs
* Contribute improvements or fixes using a [Pull Request](https://github.com/testcontainers/testcontainers-go/pulls). If you're going to contribute, thank you! Please just be sure to:
    * discuss with the authors on an issue ticket prior to doing anything big.
    * follow the style, naming and structure conventions of the rest of the project.
    * make commits atomic and easy to merge.
    * when updating documentation, please see [our guidance for documentation contributions](contributing_docs.md).
    * apply format running `go fmt`, or the shell script `./scripts/checks.sh`
    * verify all tests are passing. Build and test the project with `make test-all` to do this.

## Combining Dependabot PRs

Since we generally get quite a few Dependabot PRs, we regularly combine them into single commits. 
For this, we are using the [gh-combine-prs](https://github.com/rnorth/gh-combine-prs) extension for [GitHub CLI](https://cli.github.com/).

The whole process is as follow:

1. Check that all open Dependabot PRs did succeed their build. If they did not succeed, trigger a rerun if the cause were external factors or else document the reason if obvious.
2. Run the extension from an up-to-date local `main` branch: `gh combine-prs --query "author:app/dependabot"`
3. Merge conflicts might appear. Just ignore them, we will get those PRs in a future run.
4. Once the build of the combined PR did succeed, temporarily enable merge commits and merge the PR using a merge commit through the GitHub UI.
5. After the merge, disable merge commits again.
