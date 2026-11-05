# jira-cli (6insanes fork)

## Commands

- Build: `make build` (writes `./jira`).
- Test: `go test ./...`.
- Lint: `make lint`.
- Install locally: `install -m755 jira ~/.local/bin/jira`.

## Git

- `origin` is upstream `ankitpokhrel/jira-cli` and MUST NOT be pushed to; the fork remote is `fork` (`6insanes/jira-cli`).

## Releases

- Releases MUST follow the `jira-cli-release` skill.
- Tags MUST match `v1.8.0-6insanes.<N>`.
- GoReleaser MUST run from the `goreleaser/goreleaser:latest` docker image; installing it on the host is FORBIDDEN.
