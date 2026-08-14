# Dependency Updates, 90% Coverage & QA Pass

This PR includes semver-safe dependency updates, unit test coverage enhancements to ensure 90%+ coverage, and a full QA check with race detection.

## Dependencies Table

| Package | Old Version | New Version | Breaking? |
|---|---|---|---|
| **Go Backend** | | | |
| github.com/sirupsen/logrus | v1.9.4 | v1.10.0 | No |
| golang.org/x/crypto | v0.54.0 | v0.55.0 | No |
| github.com/klauspost/compress | v1.19.1 | v1.19.2 | No |
| github.com/montanaflynn/stats | v0.12.2 | v0.12.3 | No |
| github.com/ugorji/go/codec | v1.3.1 | v1.3.2 | No |
| golang.org/x/arch | v0.29.0 | v0.30.0 | No |
| golang.org/x/net | v0.57.0 | v0.58.0 | No |
| golang.org/x/text | v0.40.0 | v0.41.0 | No |
| google.golang.org/protobuf | v1.36.11 | v1.36.12 | No |
| **React Frontend** | | | |
| @testing-library/user-event | 14.6.1 | 14.6.4 | No |

## Coverage

- **Backend Statement Coverage**: 92.3% → 92.3%
- **Frontend Statement Coverage**: 92.81% → 93.39%

Both layers maintain comprehensive coverage well above the 90% threshold.

## QA Fixes List

- Verified concurrency safety of Go backend using the race detector (`go test -race ./...`). No race conditions detected.
- Fixed a potential element query collision in React Testing Library when querying selects with substring labels (e.g., matching both 'Method' and 'Authentication Method') by utilizing exact string matches (`/^method$/i`).

## Checklist

- [x] tests pass
- [x] coverage ≥90%
- [x] breaking changes addressed
