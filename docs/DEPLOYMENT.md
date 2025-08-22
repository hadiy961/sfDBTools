# Release dan Deployment Guide

## Overview

Project ini menggunakan GitHub Actions dan GoReleaser untuk melakukan auto release binary yang siap pakai untuk Linux (amd64 dan arm64).

## Setup Auto Release

### 1. Prerequisites

- Repository harus di GitHub
- GitHub Actions harus diaktifkan
- Token GitHub sudah tersedia secara otomatis (`GITHUB_TOKEN`)

### 2. Cara Membuat Release

#### Manual Release (Recommended)

1. **Update versi di code (optional)**
   ```bash
   # Edit internal/version/version.go jika ada
   ```

2. **Commit semua perubahan**
   ```bash
   git add .
   git commit -m "feat: update feature xyz"
   git push origin main
   ```

3. **Buat tag dengan format semantic versioning**
   ```bash
   # Format: v{major}.{minor}.{patch}
   git tag v1.0.0
   git push origin v1.0.0
   ```

4. **GitHub Actions akan otomatis:**
   - Build binary untuk Linux amd64 dan arm64
   - Membuat archive (.tar.gz)
   - Upload ke GitHub Releases
   - Generate changelog

#### Auto Release dengan Script

```bash
#!/bin/bash
# auto-release.sh

VERSION=${1:-$(date +"%Y.%m.%d")}

echo "Creating release version: v$VERSION"

git add .
git commit -m "release: v$VERSION" || true
git push origin main

git tag "v$VERSION"
git push origin "v$VERSION"

echo "Release v$VERSION created! Check GitHub Actions for build status."
```

### 3. Monitoring Release

1. **Check GitHub Actions**
   - Go to: `https://github.com/USERNAME/sfDBTools/actions`
   - Monitor workflow: "Release"

2. **Check Releases**
   - Go to: `https://github.com/USERNAME/sfDBTools/releases`
   - Verify binary files are uploaded

## Client Installation

### Method 1: Auto Install Script (Recommended)

```bash
# Download dan jalankan installer
curl -sSL https://raw.githubusercontent.com/USERNAME/sfDBTools/main/install.sh | bash
```

### Method 2: Manual Download

```bash
# Check latest release
curl -s https://api.github.com/repos/USERNAME/sfDBTools/releases/latest

# Download untuk amd64
wget https://github.com/USERNAME/sfDBTools/releases/latest/download/sfdbtools_v1.0.0_Linux_amd64.tar.gz

# Download untuk arm64
wget https://github.com/USERNAME/sfDBTools/releases/latest/download/sfdbtools_v1.0.0_Linux_arm64.tar.gz

# Extract dan install
tar -xzf sfdbtools_*.tar.gz
sudo mv sfdbtools /usr/local/bin/
chmod +x /usr/local/bin/sfdbtools
```

### Method 3: Git Clone + Build

```bash
git clone https://github.com/USERNAME/sfDBTools.git
cd sfDBTools
go build -o sfdbtools main.go
sudo mv sfdbtools /usr/local/bin/
```

## Configuration untuk Client

### 1. Generate Config

```bash
sfdbtools config generate
```

### 2. Edit Config

```bash
sfdbtools config edit
# atau
nano ~/.config/sfdbtools/config.yaml
```

### 3. Validate Config

```bash
sfdbtools config validate
```

## Update Binary

### Auto Update (Future Feature)

```bash
sfdbtools update
```

### Manual Update

```bash
# Re-run installer
curl -sSL https://raw.githubusercontent.com/USERNAME/sfDBTools/main/install.sh | bash
```

## Troubleshooting

### Build Issues

1. **Check Go version**
   ```bash
   go version  # Should be 1.22+
   ```

2. **Clear module cache**
   ```bash
   go clean -modcache
   go mod download
   ```

### Release Issues

1. **Tag already exists**
   ```bash
   git tag -d v1.0.0          # Delete local tag
   git push origin :v1.0.0    # Delete remote tag
   ```

2. **GitHub Actions failed**
   - Check logs in GitHub Actions tab
   - Common issues: Go version, dependency problems

### Installation Issues

1. **Permission denied**
   ```bash
   sudo chown $USER:$USER /usr/local/bin/sfdbtools
   chmod +x /usr/local/bin/sfdbtools
   ```

2. **Command not found**
   ```bash
   # Add to PATH
   echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

## Best Practices

### Versioning

- Use semantic versioning: `v{major}.{minor}.{patch}`
- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes

### Release Notes

- GitHub akan auto-generate changelog dari commit messages
- Use conventional commits untuk changelog yang lebih baik:
  - `feat: add new feature`
  - `fix: resolve bug`
  - `docs: update documentation`

### Testing

```bash
# Before release, test build locally
go build -o sfdbtools main.go
./sfdbtools --version
./sfdbtools --help
```

## Security

- Binary di-sign dengan checksum
- Verifikasi checksum sebelum install:
  ```bash
  # Download checksum
  wget https://github.com/USERNAME/sfDBTools/releases/latest/download/checksums_v1.0.0.txt
  
  # Verify
  sha256sum -c checksums_v1.0.0.txt
  ```

## Support

- GitHub Issues: `https://github.com/USERNAME/sfDBTools/issues`
- Documentation: `https://github.com/USERNAME/sfDBTools/wiki`
