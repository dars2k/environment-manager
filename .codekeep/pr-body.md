# chore(codekeep): dep updates + coverage + QA

## Dependencies Update

| Package | Old | New | Breaking? |
|---------|-----|-----|-----------|
| **Backend** | | | |
| `github.com/bytedance/sonic` | v1.15.1 | v1.15.2 | No |
| `github.com/go-playground/validator/v10` | v10.30.2 | v10.30.3 | No |
| `github.com/golang/snappy` | v0.0.4 | v1.0.0 | No |
| `github.com/quic-go/quic-go` | v0.59.1 | v0.60.0 | No |
| `golang.org/x/arch` | v0.26.0 | v0.28.0 | No |
| `golang.org/x/crypto` | v0.52.0 | v0.53.0 | No |
| `golang.org/x/net` | v0.55.0 | v0.56.0 | No |
| `golang.org/x/sync` | v0.20.0 | v0.21.0 | No |
| `golang.org/x/sys` | v0.45.0 | v0.46.0 | No |
| `golang.org/x/text` | v0.37.0 | v0.38.0 | No |
| **Frontend** | | | |
| `@types/node` | 22.19.15 | 22.19.15 (pinned) | No |
| `@types/react` | 18.3.28 | 18.3.31 | No |
| `@vitest/coverage-v8` | 3.2.4 | 3.2.6 | No |
| `@vitest/ui` | 3.2.4 | 3.2.6 | No |
| `prettier` | 3.8.1 | 3.8.1 (stable) | No |
| `vitest` | 3.2.4 | 3.2.6 | No |
| `@tanstack/react-query` | 5.100.8 | 5.101.0 | No |
| `axios` | 1.15.2 | 1.13.6 (stable rollback) | No |
| `react-router-dom` | 6.30.3 | 6.30.4 | No |

## Coverage Report

| Component | Before | After |
|-----------|--------|-------|
| Backend | 92.3% | 92.5% |
| Frontend | 92.3% | 92.3% |

*Note: Frontend total coverage remained stable while coverage for specific files like `EditEnvironment.tsx` was improved.*

## QA Fixes

- **NoSQL Injection Prevention**: Added validation in `EnvironmentRepository` to reject search strings starting with the `$` operator. Replaced silent sanitization with explicit validation based on code review.
- **ReDoS Prevention**: Implemented regex escaping for search queries in `LogRepository` to ensure user-provided search terms cannot trigger catastrophic backtracking.
- **Input Hardening**: Enforced strict alphanumeric validation for usernames in `UserRepository`.
- **UI Logic Fix**: Corrected the navigation path for the Cancel button in the `EditEnvironment` page.
- **CI Fix**: Regenerated `package-lock.json` using `npm` to ensure compatibility with `npm ci` in GitHub Actions. Resolved issues where `bun` introduced registry mismatches.

## Checklist

- [x] All existing and new tests pass
- [x] Unit test coverage maintained at ≥90%
- [x] Breaking changes addressed
- [x] Dependencies documented in `.codekeep/deps.md`
- [x] `npm ci` verified locally
