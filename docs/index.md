# Testcontainers

<img src="logo.png" alt="Testcontainers logo" width="1024" height="512"/>

## Test dependencies as code for your entire stack
Get lightweight and throwaway containers during your tests that let you test against any container image (database, broker, browser, etc..) using one of the several supported languages.


<p align="center"><strong>Not using Go? Here are other supported languages!</strong></p>
<div class="card-grid">
    <a href="https://testcontainers.org" class="card-grid-item"><img src="language-logos/java.svg"/>Java</a>
    <a class="card-grid-item"><img src="language-logos/go.svg"/>Go</a>
    <a href="https://dotnet.testcontainers.org/" class="card-grid-item"><img src="language-logos/dotnet.svg"/>.NET</a>
    <a href="https://testcontainers-python.readthedocs.io/en/latest/" class="card-grid-item"><img src="language-logos/python.svg"/>Python</a>
    <a href="https://github.com/testcontainers/testcontainers-node" class="card-grid-item"><img src="language-logos/javascript.svg"/><span>JavaScript<wbr>/Node.js</span></a></a>
    <a href="https://docs.rs/testcontainers/latest/testcontainers/" class="card-grid-item"><img src="language-logos/rust.svg"/>Rust</a>
</div>

## About

_Testcontainers for Go_ is a Go package that makes it simple to create and clean up container-based dependencies for
automated integration/smoke tests. The clean, easy-to-use API enables developers to programmatically define containers
that should be run as part of a test and clean up those resources when the test is done.

This project is opensource and you can have a look at the code on
[GitHub](https://github.com/testcontainers/testcontainers-go).

## GoDoc

Inline documentation and docs where the code live is crucial for us. Go has nice support for them and we provide
examples as well. Check it out at
[pkg.go.dev/github.com/testcontainers/testcontainers-go](https://pkg.go.dev/github.com/testcontainers/testcontainers-go).

## Who is using Testcontainers Go?

* [Elastic](https://www.elastic.co) - Testing of the APM Server, and E2E testing for Beats
* [Telegraf](https://www.influxdata.com/time-series-platform/telegraf/) - Integration testing the plugin-driven server agent for collecting & reporting metrics

## License

See [LICENSE](https://github.com/testcontainers/testcontainers-go/blob/main/LICENSE).

## Copyright

Copyright (c) 2018-present Gianluca Arbezzano and other authors. Check out our
[lovely contributors](https://github.com/testcontainers/testcontainers-go/graphs/contributors).
