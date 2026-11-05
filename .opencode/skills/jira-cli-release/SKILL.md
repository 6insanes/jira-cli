---
name: jira-cli-release
description: Release the 6insanes fork of jira-cli — pick the next version tag, push it to the fork, build the cross-platform assets with GoReleaser in docker, publish the GitHub release and rebuild the local binary. Trigger: any task that releases, tags, publishes or locally installs jira-cli.
---

# jira-cli release

## Repository facts

- `origin` is upstream `ankitpokhrel/jira-cli` (no write access); `fork` is
  `git@github.com:6insanes/jira-cli.git`. Pushes and releases MUST target `fork`.
- Version scheme: `v1.8.0-6insanes.<N>`, next release increments `<N>`. The current tag is
  `git tag --sort=-v:refname | head -1`.
- GoReleaser is NOT installed on the host. It MUST run from the docker image
  `goreleaser/goreleaser:latest`.
- `.goreleaser.yml` pins `release.github` to `6insanes/jira-cli` and creates a DRAFT release.
- Release assets: `jira_<version>_{linux,macOS}_*.tar.gz`, `jira_<version>_windows_x86_64.zip`,
  `checksums.txt`.

## Preconditions

- `git status --short` MUST be empty and the release commit MUST be on `main`.
- `go test ./...` and `golangci-lint run ./...` MUST pass.
- `gh auth status` MUST show the `6insanes` account with the `repo` scope.

## Steps

1. Compute and create the tag (annotated, message = tag name, like the existing ones):

```bash
NEXT=$(( $(git tag --sort=-v:refname | head -1 | sed 's/.*\.//') + 1 ))
TAG="v1.8.0-6insanes.$NEXT"
git tag -a "$TAG" -m "$TAG"
```

2. Push `main` and the tag to the fork:

```bash
git push fork main
git push fork "$TAG"
```

3. Write release notes to a file OUTSIDE the repository (a file inside the repo would dirty
   the tree). Reuse the previous release body as the template — it contains the fork summary
   and per-platform install snippets:

```bash
gh release view "$(git tag --sort=-v:refname | sed -n 2p)" -R 6insanes/jira-cli --json body -q .body > /tmp/jira-release-notes.md
# edit /tmp/jira-release-notes.md: new commit hash, new features, asset/version names
```

4. Run GoReleaser from the repository root. Publish to the fork is already pinned in the
   config; `GITHUB_TOKEN` MUST be passed from `gh`:

```bash
GITHUB_TOKEN=$(gh auth token) docker run --rm \
  -v "$PWD":/w -v /tmp:/tmp -w /w \
  -e GITHUB_TOKEN \
  goreleaser/goreleaser:latest release --clean \
  --release-notes /tmp/jira-release-notes.md
```

`goreleaser check` reports a deprecation warning for `brews` — this is expected and does not
block the release. The created release is a DRAFT (config `draft: true`).

5. Publish the draft. `gh` resolves the repository from `origin`, so `-R` MUST be passed.
   GoReleaser leaves the body empty and marks the hyphenated tag as a prerelease, so the body,
   the prerelease flag and the latest marker MUST be set here in one call:

```bash
gh release edit "$TAG" -R 6insanes/jira-cli \
  --draft=false --prerelease=false --latest \
  --notes-file /tmp/jira-release-notes.md
gh release view "$TAG" -R 6insanes/jira-cli
```

6. Build and install locally with the release version embedded:

```bash
make install VERSION="$TAG"
install -m755 "$(go env GOPATH)/bin/jira" ~/.local/bin/jira
jira version
```

`make install` puts the binary in `$(go env GOPATH)/bin`; `~/.local/bin` takes precedence in
PATH, so the copy above is the binary that runs. `make build` (`go build ./...`) only compiles
and writes NO binary.

## Checks after the release

```bash
gh release view "$TAG" -R 6insanes/jira-cli
jira issue view --help | grep -- '--comments'
jira issue view <KEY> --plain --transitions=false | head
```

- The release MUST be published (`draft: false`, `prerelease: false`), carry the body and the
  `Latest` marker.
- The Linux archive MUST contain the binary at `<tag>_linux_x86_64/bin/jira`; install snippets
  in the release body MUST point to that path (goreleaser uses `binary: bin/jira`).
- Help output MUST match the new defaults, and plain output MUST contain no ANSI escapes
  (`grep -c $'\x1b'` returns 0).

## Failure handling

- GoReleaser fails on a dirty tree: commit or stash before retrying.
- GoReleaser fails with `repository not found` or a permission error on
  `ankitpokhrel/jira-cli`: the `release.github` block in `.goreleaser.yml` is missing.
- Token missing: `gh auth login` (the release needs the `repo` scope).
- A tag already pushed with the wrong commit: delete it on the fork
  (`git push fork :refs/tags/<tag>`) and on GitHub (`gh release delete <tag> -R 6insanes/jira-cli --yes`)
  before retrying.
