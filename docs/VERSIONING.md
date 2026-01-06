# DevUp Versioning

DevUp uses automatic semantic versioning that increments with each build.

## Version Format

Versions follow [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## Automatic Version Bumping

### Default Build (Auto-Bump Patch)

Every time you run `make build`, the patch version automatically increments:

```bash
# Current version: 1.0.0
make build
# New version: 1.0.1

make build
# New version: 1.0.2
```

### Manual Version Control

You can manually bump specific version components:

```bash
# Bump patch: 1.0.0 → 1.0.1
make bump-patch

# Bump minor: 1.0.0 → 1.1.0
make bump-minor

# Bump major: 1.0.0 → 2.0.0
make bump-major
```

### Build Without Bumping

If you need to rebuild without incrementing the version:

```bash
make build-no-bump
```

## Check Current Version

```bash
# Show version from VERSION file
make version

# Show version from binary
./build/devup --version
```

## Version Storage

The version is stored in two places:

1. **`VERSION` file** - Single source of truth
2. **`cmd/root.go`** - Embedded in the binary

Both are automatically updated by the bump script.

## How It Works

### 1. VERSION File

Plain text file containing the current version:

```
1.1.0
```

### 2. Bump Script

`scripts/bump-version.sh` handles version incrementation:

- Reads current version from `VERSION` file
- Increments the appropriate component
- Updates `VERSION` file
- Updates `cmd/root.go` with new version

### 3. Build Process

```
make build
    ↓
Run bump-version.sh patch
    ↓
Update VERSION file
    ↓
Update cmd/root.go
    ↓
Build binary with new version
    ↓
Done!
```

## Version History

Track your versions in Git:

```bash
# After a significant release
git add VERSION cmd/root.go
git commit -m "Release v1.1.0"
git tag -a v1.1.0 -m "Release version 1.1.0"
git push origin v1.1.0
```

## Examples

### Regular Development

```bash
# Make changes to code
vim cmd/start.go

# Build (auto-bumps patch)
make build
# v1.0.0 → v1.0.1

# Test
./build/devup --version
# devup version 1.0.1
```

### Adding New Feature

```bash
# Before starting new feature
make bump-minor
# v1.0.5 → v1.1.0

# Develop feature...
vim cmd/newfeature.go

# Build without bumping (still developing)
make build-no-bump

# Final build when feature complete
make build
# v1.1.0 → v1.1.1
```

### Breaking Changes

```bash
# Major refactor or breaking API change
make bump-major
# v1.5.3 → v2.0.0

# Build
make build
# v2.0.0 → v2.0.1
```

## Best Practices

### 1. Let Builds Auto-Bump

For regular development, let the automatic patch bumping handle versioning:

```bash
# Just build normally
make build
```

### 2. Manual Bump for Releases

Before releasing a new feature or major change:

```bash
# New feature
make bump-minor

# Breaking change
make bump-major
```

### 3. Tag Releases

Tag significant versions in Git:

```bash
git tag -a v1.1.0 -m "Added install and setup commands"
git push origin v1.1.0
```

### 4. Document Changes

Update `CHANGELOG.md` for each minor/major version:

```markdown
## [1.1.0] - 2024-12-08

### Added
- Install command for dependency management
- Setup command for environment configuration
```

## Troubleshooting

### Version Not Updating

```bash
# Check VERSION file exists
cat VERSION

# Check script is executable
ls -la scripts/bump-version.sh
chmod +x scripts/bump-version.sh

# Manually test script
./scripts/bump-version.sh patch
```

### Binary Shows Old Version

```bash
# Rebuild completely
make clean
make build

# Check version
./build/devup --version
```

### Git Conflicts on VERSION

```bash
# Accept your version
git checkout --ours VERSION

# Or accept their version
git checkout --theirs VERSION

# Manually set version
echo "1.2.3" > VERSION
./scripts/bump-version.sh patch
```

## Commands Reference

```bash
# Version management
make version        # Show current version
make bump-patch     # Bump patch (1.0.0 → 1.0.1)
make bump-minor     # Bump minor (1.0.0 → 1.1.0)
make bump-major     # Bump major (1.0.0 → 2.0.0)

# Building
make build          # Build with auto-bump
make build-no-bump  # Build without bump

# Check
./build/devup --version               # Show binary version
cat VERSION                           # Show file version
```

## Integration with CI/CD

### Disable Auto-Bump in CI

```yaml
# .github/workflows/build.yml
- name: Build
  run: make build-no-bump
```

### Version from Git Tag

```bash
# Use git tag as version
git describe --tags --always > VERSION
make build-no-bump
```

## Semantic Versioning Guidelines

### When to Bump MAJOR (X.0.0)

- Removing commands
- Changing command behavior (breaking)
- Removing configuration options
- Incompatible changes

### When to Bump MINOR (0.X.0)

- Adding new commands
- Adding new features
- New configuration options
- Backward-compatible additions

### When to Bump PATCH (0.0.X)

- Bug fixes
- Performance improvements
- Documentation updates
- Internal refactoring (no API changes)

## Current Version

```bash
# Check current version
make version
```

Current: **1.1.0**

---

**Note**: The auto-bump feature ensures every build is uniquely versioned, making it easy to track which binary you're running and when it was built.
