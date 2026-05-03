## Summary

- **P1 – Dependency updates**: Applied all semver-safe minor/patch bumps for Go (12 packages: golang.org/x/*, go.mongodb.org/mongo-driver/v2, bytedance/sonic, gin-contrib/sse, mattn/go-isatty, pelletier/go-toml/v2) and npm (3 packages: @emotion/styled, @tanstack/react-query, axios). All major-version bumps documented in `.codekeep/deps.md` and skipped.
- **P2 – Coverage uplift**: Added 8 new test files (6 backend, 2 frontend) targeting previously uncovered branches — SSH credential paths, validateURLStrict IP edge cases, WebSocket hub/client error paths, MongoDB transaction/timeout paths, `startHealthCheckScheduler` branches, and `useNotifications` hook callbacks. Backend coverage raised from 90.0% → 91.9%; frontend from 92.23% → 92.31%.
- **P3 – QA / race-fix**: Race detector (go test -race ./...) surfaced one data race in `TestStartHealthCheckScheduler_DisabledHealthCheck` (shared `call` counter written by scheduler goroutine, read by test). Fixed by removing the racy counter; mock's built-in call tracking is sufficient. All 21 backend packages and 338 frontend tests pass clean under the race detector.

## Changes

| Area | File(s) | What |
|---|---|---|
| Go deps | `backend/go.mod`, `backend/go.sum` | 12 minor/patch upgrades |
| npm deps | `frontend/package.json`, `frontend/package-lock.json` | 3 minor/patch upgrades |
| Dep log | `.codekeep/deps.md` | Full audit of applied + skipped updates |
| Backend tests | `service/environment/service_ssh_coverage_test.go` | SSH restart/upgrade branch coverage |
| Backend tests | `service/environment/service_url_internal_test.go` | validateURLStrict IP edge cases |
| Backend tests | `websocket/hub/hub_extra_test.go` | subscribe/unsubscribe error paths, full-channel default, WritePump close |
| Backend tests | `websocket/client/client_extra_test.go` | WritePump batched-write loop, ReadPump unexpected close |
| Backend tests | `cmd/server/main_extra_test.go` | startHealthCheckScheduler branch coverage + race fix |
| Backend tests | `infrastructure/database/mongodb_txn_test.go` | Transaction abort path, NewMongoDB ping timeout |
| Frontend tests | `hooks/__tests__/useNotifications.test.ts` | Skip-already-displayed path, onExited callback, re-show after cleanup |

## Test plan

- [x] `cd backend && go test ./...` — 21/21 packages pass
- [x] `cd backend && go test -race ./...` — 0 races detected
- [x] `cd frontend && npm test -- --run` — 338/338 tests pass
- [x] `go mod verify` — all module checksums verified
- [x] No push to main/master
