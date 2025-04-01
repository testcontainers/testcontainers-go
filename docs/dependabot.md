# Staying up to date with Dependabot

It is recommended to group the Testcontainers update modules for dependabot using the following configuration:

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directories:
      - "/"
      - ...
    schedule:
      interval: "daily"
    groups:
      testcontainers-go:
        patterns:
          - "github.com/testcontainers/testcontainers-go"
          - "github.com/testcontainers/testcontainers-go/modules/*"
```

This ensures that all modules are updated at the same time, and that the PRs are grouped together. 

It also helps to avoid conflicts and makes it easier to review the changes.
