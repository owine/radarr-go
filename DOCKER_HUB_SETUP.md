# Docker Hub Integration Setup

## Overview

The GitHub Actions release workflow now uses repository variables to control Docker Hub publishing instead of checking for secrets directly in conditionals, which GitHub Actions doesn't support.

## Required Setup

### 1. Repository Variable

Go to your GitHub repository settings and set the following repository variable:

**Settings → Secrets and variables → Actions → Variables tab**

- **Name**: `ENABLE_DOCKER_HUB`
- **Value**: `true` (to enable Docker Hub publishing) or `false` (to disable)

### 2. Repository Secrets (if Docker Hub is enabled)

If you set `ENABLE_DOCKER_HUB` to `true`, you also need these secrets:

**Settings → Secrets and variables → Actions → Secrets tab**

- **Name**: `DOCKER_USERNAME`
  - **Value**: Your Docker Hub username
- **Name**: `DOCKER_PASSWORD`
  - **Value**: Your Docker Hub password or access token (recommended)

## How It Works

### Previous Issue
The workflow was using this syntax which GitHub Actions doesn't support:
```yaml
if: ${{ secrets.DOCKER_USERNAME != '' && secrets.DOCKER_PASSWORD != '' }}
```

### Current Solution
Now it uses repository variables:
```yaml
if: ${{ vars.ENABLE_DOCKER_HUB == 'true' }}
```

## Workflow Behavior

### When `ENABLE_DOCKER_HUB` is `true`
- Logs into Docker Hub using the provided secrets
- Builds and pushes images to Docker Hub
- Includes Docker Hub information in release summaries

### When `ENABLE_DOCKER_HUB` is `false` or not set
- Skips Docker Hub login, metadata extraction, and pushing
- Only publishes to GitHub Container Registry
- Omits Docker Hub references from release summaries

## Benefits

1. **Explicit Control**: Clear variable to enable/disable Docker Hub publishing
2. **GitHub Actions Compliance**: Uses supported conditional syntax
3. **Flexible Deployment**: Can be toggled without workflow changes
4. **Better Error Handling**: Clearer indication when Docker Hub is intentionally disabled

## Migration from Previous Setup

If you had the workflow running before, you need to:

1. Set the `ENABLE_DOCKER_HUB` repository variable to `true`
2. Ensure your `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets are still configured
3. The workflow will now work correctly without validation errors

## Security Notes

- Repository variables are visible to all repository collaborators
- Use `ENABLE_DOCKER_HUB=false` if you don't want to publish to Docker Hub
- Secrets remain secure and are only accessible during workflow execution
- Consider using Docker Hub access tokens instead of passwords
