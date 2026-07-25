## Summary

- **P1 – Dependency updates**: Applied semver-safe minor/patch bumps for Go (7 direct/indirect packages via `go get -u ./... && go mod tidy`) and npm (`vite` 7.3.2 → 7.3.6, plus transitive advisories closed via `npm audit fix`, including a **critical** Vitest UI arbitrary-file-read CVE and a moderate `react-router`/`react-router-dom` open-redirect CVE). All major-version bumps documented and skipped in `.codekeep/deps.md`.
- **P2 – Coverage uplift**: Frontend overall coverage 92.29% → 93.27%. Added a dedicated test file for `EditEnvironment.tsx`'s submit flow (password/private-key branches, success/error handling) taking that file from 64.28% → 100%. Added reset-password validation/failure/success tests and a delete-failure test for `Users.tsx`, taking it from 85.6% → 93.43%. Backend already sat at 92.3% aggregate with no regressions from the dependency bumps (`go test ./... -race` all green).
- **P3 – QA**: While adding the `Users.tsx` tests, found and fixed a **real pre-existing bug**: several existing tests queried `[data-testid="KeyIcon"]` to find the reset-password icon button, but MUI's `VpnKey` icon actually renders `data-testid="VpnKeyIcon"` — the old selector never matched, so those tests were silently vacuous (guarded by an `if (length > 0)` that always evaluated false). Fixed the selector everywhere it's used. Also noticed the frontend has no ESLint config file at all despite a `lint` script and ESLint deps — pre-existing, documented, not fixed here (out of scope).

## Dependency table

### Go (backend)

| Package | Old | New | Breaking? |
|---|---|---|---|
| github.com/go-playground/validator/v10 | v10.30.2 | v10.30.3 | No |
| golang.org/x/crypto | v0.52.0 | v0.54.0 | No |
| github.com/bytedance/sonic (indirect) | v1.15.1 | v1.15.2 | No |
| github.com/gabriel-vasile/mimetype (indirect) | v1.4.13 | v1.4.15 | No |
| github.com/golang/snappy (indirect) | v0.0.4 | v1.0.0 | No |
| github.com/klauspost/compress (indirect) | v1.18.6 | v1.19.1 | No |
| github.com/klauspost/cpuid/v2 (indirect) | v2.3.0 | v2.4.0 | No |
| github.com/leodido/go-urn (indirect) | v1.4.0 | v1.5.0 | No |
| github.com/mattn/go-isatty (indirect) | v0.0.22 | v0.0.24 | No |
| github.com/montanaflynn/stats (indirect) | v0.9.0 | v0.12.2 | No |
| github.com/pelletier/go-toml/v2 (indirect) | v2.3.1 | v2.4.3 | No |
| github.com/quic-go/quic-go (indirect) | v0.59.1 | v0.61.0 | No |
| go.mongodb.org/mongo-driver/v2 (indirect) | v2.6.0 | v2.8.0 | No |
| golang.org/x/arch (indirect) | v0.26.0 | v0.29.0 | No |
| golang.org/x/net (indirect) | v0.55.0 | v0.57.0 | No |
| golang.org/x/sync (indirect) | v0.20.0 | v0.22.0 | No |
| golang.org/x/sys (indirect) | v0.45.0 | v0.47.0 | No |
| golang.org/x/text (indirect) | v0.37.0 | v0.40.0 | No |

### npm (frontend)

| Package | Old | New | Breaking? |
|---|---|---|---|
| vite | 7.3.2 | 7.3.6 | No (patch; was exact-pinned) |
| esbuild, @typescript-eslint/*, jest-resolve-dependencies, @jest/reporters, react-router, react-router-dom | — | in-range via `npm audit fix` | No |

### Major bumps documented, not applied

See `.codekeep/deps.md` for the full table and reasoning (mongo-driver v1→v2, React 18→19, MUI 5→9, react-router-dom 6→7, @reduxjs/toolkit 1→2, date-fns 2→4, recharts 2→3).

## Coverage: before → after

| Suite | Before | After |
|---|---|---|
| Backend (`go test ./... -coverprofile`) | 92.3% | 92.3% (unchanged; no backend test gaps were closed this round) |
| Frontend (`vitest run --coverage`) | 92.29% | **93.27%** ✅ |
| `frontend/src/pages/EditEnvironment.tsx` | 64.28% | **100%** ✅ |
| `frontend/src/pages/Users.tsx` | 85.6% | **93.43%** ✅ |

Both suites' aggregate coverage sits comfortably above the 90% target. Remaining sub-90% spots (`cmd/server` main() entrypoint at the backend, a handful of dialog/form components on the frontend) are either integration-level bootstrap code or files already covered above the aggregate threshold; not chased further this round to keep the change scoped.

## QA fixes

1. **Vacuous test selector fixed** (`frontend/src/pages/__tests__/Users.test.tsx`): `[data-testid="KeyIcon"]` never matched (the real MUI `VpnKey` icon renders `data-testid="VpnKeyIcon"`), so the "opens reset password dialog" and related assertions were silently no-ops via an `if (length > 0)` guard. Corrected the selector; the previously-untested reset-password dialog interactions now actually execute and are covered by new tests (validation, failure, success paths).
2. Added a new `EditEnvironment.submit.test.tsx` covering the previously-untested `handleSubmit`/`updateMutation` flow (password branch, private-key branch, no-credentials branch, server-error branch, generic-error fallback).

## Checklist

- [x] Tests pass — `go test ./... -race` (20/20 backend packages green)
- [x] Tests pass — `vitest run` (31/31 frontend files, 348 tests)
- [x] Coverage ≥90% (backend 92.3%, frontend 93.27%)
- [x] Breaking changes addressed (none applied; all documented in `.codekeep/deps.md`)
- [x] No push to main/master
