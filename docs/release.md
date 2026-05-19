# Release Checklist

Use this checklist before tagging a release.

Releases use semantic version tags in the form `v*.*.*`, for example
`v0.1.0`.

## API And Documentation

- Review exported APIs with `go doc .`.
- Confirm README examples match the current public API.
- Update `CHANGELOG.md`.
- Check `docs/compatibility.md` for any compatibility-policy changes.
- Confirm deprecated exported fields still have clear replacement guidance.

## Correctness

```sh
make test
make race
make vet
make staticcheck
make perft
make fuzz-smoke
```

## Benchmarks

```sh
make bench-smoke
make bench-snapshot
```

For performance-sensitive releases, capture a before/after comparison with
`benchstat` and keep the notable output under `docs/benchmarks/`.

## Tagging

- Confirm the working tree contains only intended changes.
- Confirm `CHANGELOG.md` has a section for the release version.
- Create an annotated semver tag:

```sh
git tag -a v0.1.0 -m "Release v0.1.0"
```

- Push the branch and tag after CI is green:

```sh
git push origin main
git push origin v0.1.0
```

Pushing a `v*.*.*` tag creates a GitHub Release.
