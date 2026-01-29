# TERA Release Workflow

This document describes how to release new versions of TERA.

## Prerequisites

- [GoReleaser](https://goreleaser.com/install/) installed locally (for testing)
- Push access to the repository
- Git tags follow [Semantic Versioning](https://semver.org/)

## Version Format

| Type              | Format        | Example       | When to use                        |
| ----------------- | ------------- | ------------- | ---------------------------------- |
| Release Candidate | `vX.Y.Z-rc.N` | `v1.0.0-rc.1` | Pre-release testing                |
| Stable            | `vX.Y.Z`      | `v1.0.0`      | Production release                 |
| Patch             | `vX.Y.Z`      | `v1.0.1`      | Bug fixes                          |
| Minor             | `vX.Y.Z`      | `v1.1.0`      | New features (backward compatible) |
| Major             | `vX.Y.Z`      | `v2.0.0`      | Breaking changes                   |

## Release Steps

### 1. Prepare the Release

```sh
# Test locally
make clean && make lint && make build && ./tera
./tera --version
```

### 2. Test GoReleaser Locally (Optional but Recommended)

```sh
# Dry run - shows what would be built without publishing
goreleaser release --snapshot --clean

# Check the dist/ folder for generated binaries
ls -la dist/
```

### 3. Create and Push the Tag

```sh
# Create annotated tag
git tag -a v1.0.0-rc.x -m "Release v1.0.0-rc.x"

# Push the tag (this triggers the GitHub Action)
git push origin v1.0.0-rc.x
```

### 4. Monitor the Release

1. Go to **GitHub → Actions** tab
2. Watch the "Release" workflow
3. Once complete, check **Releases** page for the new release

### 5. Verify the Release

```sh
# Test installation via Go
go install github.com/shinokada/tera/cmd/tera@v1.0.0-rc.1
tera --version
# Should show: TERA v1.0.0-rc.1

# Or download from GitHub Releases and test
```

## Quick Reference

### Release a New Version

```sh
# For release candidate
git tag -a v1.0.0-rc.1 -m "Release v1.0.0-rc.1"
git push origin v1.0.0-rc.1

# For stable release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# For patch release
git tag -a v1.0.1 -m "Release v1.0.1"
git push origin v1.0.1
```

### Delete a Tag (if something went wrong)

```sh
# Delete local tag
git tag -d v1.0.0-rc.1

# Delete remote tag
git push origin --delete v1.0.0-rc.1

# Also delete the GitHub Release manually if created
```

### List All Tags

```sh
git tag -l "v*" --sort=-version:refname
```

## Release Checklist

- [ ] All tests pass (`go test ./...`)
- [ ] Code builds successfully (`go build ./cmd/tera`)
- [ ] README is up to date
- [ ] CHANGELOG updated (if you maintain one)
- [ ] GoReleaser dry run successful (`goreleaser release --snapshot --clean`)
- [ ] Tag created and pushed
- [ ] GitHub Action completed successfully
- [ ] Release page has correct binaries and notes
- [ ] `go install ...@<version>` works

## Troubleshooting

### GoReleaser fails

```sh
# Check configuration
goreleaser check

# Run with debug output
goreleaser release --snapshot --clean --debug
```

### Wrong version in binary

The version comes from the git tag. Ensure:
1. Tag follows `vX.Y.Z` format
2. Tag is annotated (`git tag -a`)
3. GoReleaser ldflags include `-X main.version={{.Version}}`

### GitHub Action fails

1. Check Actions tab for error logs
2. Ensure `GITHUB_TOKEN` has write permissions
3. Verify `.goreleaser.yaml` syntax

## Future Enhancements

When ready to expand distribution:

1. **Homebrew Tap**: Create `shinokada/homebrew-tap` repository, then uncomment `brews` section in `.goreleaser.yaml`
2. **Scoop Bucket**: Create `shinokada/scoop-bucket` repository, then uncomment `scoops` section
3. **Snap/Flatpak**: Add additional configuration as needed
