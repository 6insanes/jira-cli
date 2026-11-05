# jira-cli (6insanes fork)

## Commands

- Test: `go test ./...`.
- Lint: `make lint`.
- Install locally: `make install VERSION=$(git describe --tags --exact-match)` then
  `install -m755 "$(go env GOPATH)/bin/jira" ~/.local/bin/jira`. `make build` only compiles and
  writes no binary.

## Git

- `origin` is upstream `ankitpokhrel/jira-cli` and MUST NOT be pushed to; the fork remote is `fork` (`6insanes/jira-cli`).

## Releases

- Releases MUST follow the `jira-cli-release` skill.
- Tags MUST match `v1.8.0-6insanes.<N>`.
- GoReleaser MUST run from the `goreleaser/goreleaser:latest` docker image; installing it on the host is FORBIDDEN.
