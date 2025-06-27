# GitHub Actions CI/CD Pipeline

This directory contains GitHub Actions workflows for automated testing, building, and releasing the Environment Manager application.

## Workflows

### 1. CI/CD Pipeline (`ci-cd.yml`)

The main pipeline that runs on commits to `master`/`main` branches and pull requests.

#### Optimizations Implemented:

1. **Alpine-based containers**: Uses lightweight Alpine Linux containers for test jobs
   - Backend tests run in `golang:1.21-alpine` 
   - Frontend tests run in `node:20-alpine`
   - Reduces overhead compared to full Ubuntu runners

2. **Caching strategies**:
   - Go modules and build cache for backend
   - NPM cache and node_modules for frontend
   - Docker layer caching with GitHub Actions cache
   - Reduces build times significantly

3. **Concurrency control**:
   - Cancels in-progress runs for the same branch
   - Saves resources and speeds up feedback loop

4. **Artifact retention**:
   - Coverage reports kept for only 1 day
   - Reduces storage costs

5. **Optimized npm commands**:
   - Uses `npm ci` with `--prefer-offline --no-audit`
   - Faster installs by skipping audit and using cache

#### Job Structure:

1. **test-backend**: Runs Go tests with race detection and coverage
2. **test-frontend**: Runs npm tests with coverage
3. **build-and-push**: Builds and pushes Docker images to GitHub Container Registry
4. **integration-test**: Runs full stack tests
5. **release**: Creates GitHub releases with changelogs

### 2. Deployment Notification (`deploy-notification.yml`)

Sends notifications when the CI/CD pipeline completes successfully on the master branch.

## Best Practices

1. **Container-based testing**: Running tests in containers ensures consistency and reduces setup time
2. **Parallel execution**: Backend and frontend tests run in parallel
3. **Docker buildx**: Uses buildx for advanced caching and multi-platform builds
4. **GitHub Container Registry**: Uses ghcr.io for free image hosting with the repository
5. **Automated versioning**: Increments version numbers automatically
6. **Security**: Uses GITHUB_TOKEN for authentication (no separate secrets needed)

## Performance Comparison

| Approach | Estimated Time | Resources |
|----------|---------------|-----------|
| Ubuntu runners (default) | ~5-7 minutes | Standard VMs |
| Alpine containers | ~3-5 minutes | Lightweight containers |
| With caching | ~2-3 minutes | Cached dependencies |

## Future Optimizations

1. **Matrix builds**: Test against multiple Go/Node versions
2. **Self-hosted runners**: For even more control and performance
3. **Dependency caching service**: Consider using a dedicated cache service
4. **Build-only on changes**: Skip unchanged services using path filters

## Debugging

If a workflow fails:
1. Check the logs in the Actions tab
2. Integration test failures show docker-compose logs
3. Use `act` locally to test workflows: `act -j test-backend`

## Requirements

- Go 1.21+
- Node.js 20+
- Docker and Docker Compose
- GitHub repository with Actions enabled
