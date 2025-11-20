# Release Process

This project uses [GoReleaser](https://goreleaser.com) with GitHub Actions to automatically build and publish binaries when a new tag is pushed.

## How to Create a Release

1. **Update the version** (if needed):
   - Update any version references in code or documentation

2. **Create and push a tag**:
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

3. **GitHub Actions will automatically**:
   - Build binaries for all supported platforms:
     - Linux (amd64, arm64)
     - macOS (amd64, arm64)
     - Windows (amd64, arm64)
   - Create a GitHub Release with all binaries attached
   - Generate release notes
   - Include SHA256 checksums for verification

## Supported Platforms

- **Linux**: `que-linux-amd64.tar.gz`, `que-linux-arm64.tar.gz`
- **macOS**: `que-darwin-amd64.tar.gz`, `que-darwin-arm64.tar.gz`
- **Windows**: `que-windows-amd64.zip`, `que-windows-arm64.zip`

## Tag Naming Convention

- Use semantic versioning: `v1.0.0`, `v0.2.1`, etc.
- Tags starting with `v` will trigger the release workflow
- Pre-release tags (containing `-`) will be marked as pre-releases: `v1.0.0-beta.1`

## Manual Release (if needed)

If you need to create a release manually:

1. Go to the [Releases page](https://github.com/jenian/que/releases)
2. Click "Draft a new release"
3. Select or create a tag
4. Upload the binaries manually

## Verifying Binaries

Each release includes a `checksums.txt` file with SHA256 checksums. Verify a binary:

```bash
# Linux/macOS
sha256sum -c checksums.txt

# macOS (alternative)
shasum -a 256 -c checksums.txt
```

