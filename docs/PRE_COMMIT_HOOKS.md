# Pre-commit Hooks Documentation

This document explains the comprehensive pre-commit hooks setup for Radarr Go, providing automated code quality checks and formatting before commits.

## Overview

Our pre-commit hooks are organized into categories and provide comprehensive linting for all file types in the project:

- **Go Development**: Formatting, linting, imports, modules, and tests
- **Frontend Development**: TypeScript/React linting and compilation checks
- **Configuration & Documentation**: YAML, JSON, Markdown validation
- **Shell Scripts**: ShellCheck validation
- **General File Quality**: Formatting, security, and git checks

## Installation and Setup

### Automatic Setup
```bash
# Complete development environment setup (includes pre-commit)
./scripts/dev-setup.sh

# Or install pre-commit manually
pip install pre-commit
pre-commit install
```

### Manual Setup
```bash
# Install pre-commit hooks
pre-commit install --install-hooks

# Update hook versions
pre-commit autoupdate
```

## Hook Categories

### 1. Go Development Hooks

#### Go Format (`go-fmt`)
- **Purpose**: Format Go code with `gofmt -s`
- **Auto-fix**: Yes
- **Files**: `*.go`
- **Command**: `gofmt -s`

#### Go Imports (`go-imports`)  
- **Purpose**: Organize Go imports with `goimports`
- **Auto-fix**: Yes
- **Files**: `*.go`
- **Command**: `goimports`

#### Go Mod Tidy (`go-mod-tidy`)
- **Purpose**: Tidy and verify Go modules
- **Auto-fix**: Yes
- **Files**: `go.mod`, `go.sum`
- **Command**: `go mod tidy`

#### Go Lint (`golangci-lint`)
- **Purpose**: Comprehensive Go linting
- **Auto-fix**: Partial
- **Files**: `*.go`
- **Config**: `.golangci.yml`
- **Command**: `golangci-lint run --config=.golangci.yml`

#### Go Unit Tests (`go-unit-tests`)
- **Purpose**: Run Go unit tests
- **Stage**: Manual only (use `--hook-stage manual`)
- **Timeout**: 2 minutes
- **Command**: `go test -timeout=2m -short`

### 2. Frontend Development Hooks

#### Frontend ESLint (`eslint-frontend`)
- **Purpose**: Lint TypeScript/React files with ESLint
- **Auto-fix**: Yes
- **Files**: `web/frontend/src/**/*.{ts,tsx}`
- **Config**: `web/frontend/eslint.config.js`
- **Command**: `cd web/frontend && npx eslint --fix`

#### TypeScript Check (`typescript-check`)
- **Purpose**: Check TypeScript compilation without emitting files
- **Auto-fix**: No
- **Files**: `web/frontend/src/**/*.{ts,tsx}`
- **Config**: `web/frontend/tsconfig.json`
- **Command**: `cd web/frontend && npx tsc --noEmit`

### 3. Configuration & Documentation Hooks

#### YAML Lint (`yamllint`)
- **Purpose**: Validate and lint YAML files
- **Auto-fix**: No
- **Files**: `*.yml`, `*.yaml`
- **Config**: `.yamllint.yml`
- **Excludes**: `node_modules`, `radarr-source`, GitHub workflows
- **Command**: `yamllint -c .yamllint.yml`

#### JSON Syntax Check (`check-json`)
- **Purpose**: Validate JSON file syntax
- **Auto-fix**: No
- **Files**: `*.json`
- **Excludes**: `node_modules`, `*.example.json`, `tsconfig*.json`
- **Command**: `python -m json.tool`

#### Markdown Lint (`markdownlint`)
- **Purpose**: Lint Markdown files for consistency
- **Auto-fix**: Yes (some rules)
- **Files**: `*.md`
- **Config**: `.markdownlint.json`
- **Excludes**: `node_modules`, `radarr-source`, `CHANGELOG.md`
- **Command**: `markdownlint --config .markdownlint.json --fix`

### 4. Shell Script Hooks

#### Shell Script Lint (`shellcheck`)
- **Purpose**: Lint shell scripts with ShellCheck
- **Auto-fix**: No (provides suggestions)
- **Files**: `*.sh`, `*.bash`
- **Args**: `-x -e SC1091` (enable external sourcing, ignore SC1091)
- **Excludes**: `node_modules`, `radarr-source`
- **Command**: `shellcheck -x -e SC1091`

### 5. General File Hooks

#### Trim Trailing Whitespace (`trailing-whitespace`)
- **Purpose**: Remove trailing whitespace from files
- **Auto-fix**: Yes
- **Special**: Preserves Markdown line breaks
- **Excludes**: Binary files, patches, diff files

#### Fix End of Files (`end-of-file-fixer`)
- **Purpose**: Ensure files end with a newline
- **Auto-fix**: Yes
- **Excludes**: Binary files (images, fonts, etc.)

#### Fix Line Endings (`mixed-line-ending`)
- **Purpose**: Normalize line endings to LF
- **Auto-fix**: Yes
- **Args**: `--fix=lf`
- **Excludes**: Windows batch files (`.bat`, `.cmd`)

#### Check Large Files (`check-added-large-files`)
- **Purpose**: Prevent large files from being committed
- **Limit**: 1000KB
- **Auto-fix**: No (manual review required)

#### Security Checks
- **Check Merge Conflicts**: Detect merge conflict markers
- **Check Case Conflicts**: Detect case conflicts in filenames
- **Check Symlinks**: Validate symbolic links
- **Check VCS Permalinks**: Validate VCS permalinks
- **Forbid Submodules**: Prevent new git submodules

## Usage Examples

### Run All Hooks
```bash
# Run on staged files (normal commit behavior)
pre-commit run

# Run on all files
pre-commit run --all-files

# Run specific hook
pre-commit run yamllint
pre-commit run eslint-frontend
```

### Run Specific Hook Categories
```bash
# Go linting only
pre-commit run go-fmt go-imports golangci-lint

# Frontend linting only  
pre-commit run eslint-frontend typescript-check

# Documentation linting
pre-commit run yamllint markdownlint

# All formatting hooks
pre-commit run trailing-whitespace end-of-file-fixer mixed-line-ending
```

### Skip Hooks
```bash
# Skip specific hooks
SKIP=go-unit-tests git commit -m "Quick fix"

# Skip all hooks (emergency commits only)
git commit --no-verify -m "Emergency fix"

# Skip hooks for specific files
git commit -m "Update docs" -- README.md
```

### Manual Testing Mode
```bash
# Run tests manually (not on every commit)
pre-commit run --hook-stage manual go-unit-tests
```

## Performance Optimization

### Hook Execution Order
Hooks are organized for optimal performance:
1. **Fast formatting** (go-fmt, trailing-whitespace) run first
2. **Medium linting** (golangci-lint, eslint) run second  
3. **Slow operations** (tests) run last or manually only

### File Filtering
Each hook only runs on relevant file types:
- Go hooks: `*.go` files only
- Frontend hooks: `web/frontend/src/**/*.{ts,tsx}` only
- YAML hooks: `*.yml`, `*.yaml` files only
- etc.

### Exclusion Patterns
Strategic exclusions prevent unnecessary processing:
- `node_modules/` directories
- `radarr-source/` legacy code
- Build artifacts (`dist/`, `build/`)
- Binary files (images, fonts)

## Configuration Files

### Pre-commit Config
- **File**: `.pre-commit-config.yaml`
- **Purpose**: Defines all hooks and their configuration
- **Updates**: Use `pre-commit autoupdate` to update hook versions

### Tool-Specific Configs
- **Go**: `.golangci.yml` - golangci-lint configuration
- **YAML**: `.yamllint.yml` - yamllint rules
- **Markdown**: `.markdownlint.json` - markdownlint rules
- **TypeScript**: `web/frontend/eslint.config.js` - ESLint configuration

## Integration with Makefile

Pre-commit hooks complement Makefile targets:

```bash
# Manual linting (same as pre-commit hooks)
make lint-all          # All linters
make lint-go           # Go linting (golangci-lint)
make lint-frontend     # Frontend linting (ESLint)
make lint-yaml         # YAML linting (yamllint)  
make lint-markdown     # Markdown linting
make lint-shell        # Shell script linting

# Automatic formatting
make fmt               # Go formatting (gofmt + goimports)
```

## Troubleshooting

### Common Issues

#### Hook Installation Failed
```bash
# Clear cache and reinstall
pre-commit clean
pre-commit install --install-hooks
```

#### golangci-lint Version Issues  
```bash
# Update to stable version
pre-commit autoupdate --repo https://github.com/golangci/golangci-lint
```

#### Frontend Hooks Not Running
```bash
# Ensure Node.js dependencies are installed
cd web/frontend
npm install
```

#### YAML/Markdown Linting Errors
```bash  
# Install Python linting tools
pip install yamllint
npm install -g markdownlint-cli
```

### Performance Issues

#### Slow Hook Execution
```bash
# Run hooks in parallel (default behavior)
# Or skip slow hooks temporarily
SKIP=go-unit-tests,typescript-check git commit -m "Fast commit"
```

#### Large Repository Issues
```bash
# Run hooks only on changed files
pre-commit run --files file1.go file2.ts
```

## Best Practices

### For Developers

1. **Install hooks immediately** after cloning the repository
2. **Run `make fmt`** before committing to minimize hook failures  
3. **Use `pre-commit run --all-files`** when updating hook configurations
4. **Keep commits small** for faster hook execution
5. **Fix linting issues** rather than skipping hooks

### For Maintainers

1. **Update hook versions regularly** with `pre-commit autoupdate`
2. **Test hook changes** with `pre-commit run --all-files`
3. **Document new hooks** in this file
4. **Monitor hook performance** and optimize as needed
5. **Keep exclusion patterns minimal** but effective

### Emergency Procedures

1. **Critical hotfixes**: Use `--no-verify` for emergency commits
2. **Hook failures in CI**: Fix linting issues in follow-up commits
3. **Broken hooks**: Temporarily disable problematic hooks in config
4. **Performance issues**: Temporarily skip expensive hooks

## Hook Maintenance

### Regular Updates
```bash
# Update all hooks to latest versions
pre-commit autoupdate

# Test updates
pre-commit run --all-files

# Commit updates
git add .pre-commit-config.yaml
git commit -m "chore: update pre-commit hooks"
```

### Adding New Hooks
1. **Add to `.pre-commit-config.yaml`**
2. **Update this documentation**  
3. **Test with `pre-commit run --all-files`**
4. **Add corresponding Makefile target if needed**

### Hook Configuration Changes
1. **Update tool config files** (`.golangci.yml`, `.yamllint.yml`, etc.)
2. **Test changes** with `pre-commit run --all-files`
3. **Document changes** in commit messages and this file

## Integration with CI/CD

Pre-commit hooks complement CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Run pre-commit hooks
  run: |
    pip install pre-commit
    pre-commit run --all-files
```

This ensures the same quality checks run locally and in CI, providing fast feedback to developers while maintaining consistency.