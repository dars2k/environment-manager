## chore(codekeep): dep updates + coverage ≥90% + QA pass

### Phase 1 — Dependency Updates

#### Go (backend)

| Package | Old | New | Breaking? |
|---|---|---|---|
| golang.org/x/crypto | v0.50.0 | v0.52.0 | No |
| golang.org/x/net | v0.53.0 | v0.55.0 | No |
| golang.org/x/text | v0.36.0 | v0.37.0 | No |
| golang.org/x/sys | v0.43.0 | v0.45.0 | No |
| golang.org/x/arch | v0.26.0 | v0.27.0 | No |
| github.com/go-playground/validator/v10 | v10.30.2 | v10.30.3 | No |

#### npm (frontend)

| Package | Old | New | Breaking? |
|---|---|---|---|
| axios | ^1.15.2 | ^1.16.1 | No |
| ws (transitive) | vulnerable | patched | No — CVE GHSA-58qx-3vcg-4xpx (uninitialized memory disclosure) |
| yaml (transitive) | vulnerable | patched | No — CVE GHSA-48c2-rrv3-qjmp (stack overflow) |

### Phase 2 — Test Coverage

| Scope | Before | After |
|---|---|---|
| Backend overall | 91.9% | 92.1% |
| Frontend statements | ~88% | 93.09% |
| `websocket/client` | 85.1% | 91.0% |
| `websocket/hub` | 89.2% | 90.2% |
| `EditEnvironment.tsx` | 64.28% | 100% |
| `EnvironmentForm.tsx` | 83.21% | 85.86% |
| `Users.tsx` | 85.60% | 86.11% |

New test files added:
- `backend/internal/infrastructure/database/mongodb_connect_test.go` — `NewMongoDB` error paths (invalid URI scheme, empty URI)
- `backend/internal/websocket/client/client_coverage2_test.go` — `sendMessage` marshal error, ReadPump unexpected close, WritePump NextWriter error
- `backend/internal/websocket/hub/hub_coverage2_test.go` — ReadPump unexpected close, WritePump WriteJSON error
- `frontend/src/pages/__tests__/EditEnvironmentCoverage.test.tsx` — all `EditEnvironment` mutation paths (password, key, error, fallback error, no-metadata)

Extended test files:
- `frontend/src/pages/__tests__/Users.test.tsx` — passwords-mismatch error, reset-password API error, delete-user API error
- `frontend/src/components/environments/__tests__/EnvironmentForm.test.tsx` — SSH upgrade commands section render and input

### Phase 3 — QA Review

| Check | Result |
|---|---|
| `go test -race ./...` | Clean — 0 races |
| `go vet ./...` | Clean — 0 issues |
| Frontend tests | 348 tests pass |
| Logic errors | None found |
| Brittle tests | None found |

QA fixes: **0** — all checks passed clean.

### Checklist

- [x] All backend tests pass (`go test ./...`)
- [x] All frontend tests pass (`npm test`)
- [x] Coverage ≥ 90% (backend 92.1%, frontend 93.09%)
- [x] Race detector clean (`go test -race`)
- [x] Static analysis clean (`go vet`)
- [x] No breaking dependency changes
- [x] CVE-affected transitive packages patched
