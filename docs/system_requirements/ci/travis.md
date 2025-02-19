# Travis

To run Testcontainers on TravisCI, docker needs to be installed. The configuration below
is the minimal required config.

```yaml
language: go
go:
- 1.x
- "1.23"

services:
- docker

script: go test ./... -v
```
