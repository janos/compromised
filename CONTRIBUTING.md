# How to contribute

We'd love to accept your patches and contributions to this project. There are just a few small guidelines you need to follow.

1. Code should be `go fmt` formatted.
2. Exported types, constants, variables and functions should be documented.
3. Changes must be covered with tests.
4. All tests must pass constantly by running the `make` command.

## Versioning

Compromised service follows semantic versioning. New functionality should be accompanied by increment to the minor version number.

## Releasing

Any code which is complete, tested, reviewed, and merged to master can be released.

Releasing a new version is automated with goreleaser and GitHub Actions.

To release, only a tag with semantic version and a prefix `v` should be pushed to the repository on GitHub, e.g. `v1.25.0`.
