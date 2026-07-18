## Summary

- **P1 – Dependency updates**: Applied semver-safe patch/minor bumps for Go (19 packages, mostly `golang.org/x/*` plus a handful of indirect deps) and npm (`@tanstack/react-query`). All major-version bumps documented and skipped in `.codekeep/deps.md`; two npm packages (`axios`, `prettier`) were left untouched because the package registry mirror available in this environment reports no version newer than what's already installed.
- **P2 – Coverage check**: Re-ran full coverage. No regressions from the dependency bumps. Coverage is unchanged from the last audit: all backend packages remain ≥90% except `cmd/server` (39.0%) and `internal/infrastructure/database` (85.0%), both already documented as blocked on hard infrastructure constraints (see below) rather than missing test effort. Frontend remains at 92.31% overall. No new tests were added this cycle since there was no new untested surface — this was a dependency-only change.
- **P3 – QA**: `go vet ./...` clean, no logic changes were introduced (dependency version bumps only), all 21 backend packages and all 30 frontend test files pass. Confirmed the two directly-used bumped packages (`golang.org/x/crypto` for bcrypt/SSH) still pass their existing test suites.

## Dependency table

### Go (backend)

| Package | Old | New | Breaking? |
|---|---|---|---|
| github.com/go-playground/validator/v10 | v10.30.2 | v10.30.3 | No |
| github.com/bytedance/sonic | v1.15.1 | v1.15.2 | No |
| github.com/golang/protobuf | v1.5.0 | v1.5.4 | No |
| github.com/google/gofuzz | v1.0.0 | v1.2.0 | No |
| github.com/klauspost/compress | v1.18.6 | v1.19.0 | No |
| github.com/klauspost/cpuid/v2 | v2.3.0 | v2.4.0 | No |
| github.com/mattn/go-isatty | v0.0.22 | v0.0.23 | No |
| github.com/pelletier/go-toml/v2 | v2.3.1 | v2.4.3 | No |
| github.com/rogpeppe/go-internal | v1.10.0 | v1.15.0 | No |
| github.com/yuin/goldmark | v1.4.13 | v1.8.4 | No |
| golang.org/x/arch | v0.26.0 | v0.29.0 | No |
| golang.org/x/crypto | v0.52.0 | v0.54.0 | No |
| golang.org/x/mod | v0.35.0 | v0.38.0 | No |
| golang.org/x/net | v0.55.0 | v0.57.0 | No |
| golang.org/x/sync | v0.20.0 | v0.22.0 | No |
| golang.org/x/sys | v0.45.0 | v0.47.0 | No |
| golang.org/x/term | v0.43.0 | v0.45.0 | No |
| golang.org/x/text | v0.37.0 | v0.40.0 | No |
| golang.org/x/tools | v0.44.0 | v0.48.0 | No |

Skipped (major/indirect, documented in `.codekeep/deps.md`): `golang.org/x/snappy` (v0.0.4→v1.0.0), `montanaflynn/stats` (v0.9.0→v0.12.2), `quic-go/quic-go` (v0.59.1→v0.60.0), `mongo-driver/v2` (v2.6.0→v2.8.0, transitive only).

### npm (frontend)

| Package | Old | New | Breaking? |
|---|---|---|---|
| @tanstack/react-query | 5.100.8 | 5.101.2 | No |

`axios` and `prettier` left unchanged — see `.codekeep/deps.md` for why.

## Coverage: before → after

| Package | Coverage |
|---|---|
| `internal/infrastructure/database` | 85.0% (unchanged) ⚠️ |
| `cmd/server` | 39.0% (unchanged) ⚠️ |
| All other backend packages | ≥90% ✅ |
| Frontend overall | 92.31% (unchanged) ✅ |

**Why `database` and `cmd/server` are capped (unchanged from prior audit):**
- `database.NewMongoDB`/`Transaction` success paths and `StartTransaction`/`CommitTransaction`/`AbortTransaction` require a live MongoDB (replica set for transactions); the mtest mock and CI sandbox have no reachable MongoDB, so these branches can't execute in this environment.
- `cmd/server.main()` calls `logger.Fatal()` on MongoDB failure → `os.Exit()`, unrunnable in unit tests without a DI refactor of `main()`, which is out of scope for a dependency maintenance pass.

## QA notes

- No source logic was changed this cycle — only dependency versions and `.codekeep/deps.md`.
- `go vet ./...` and `go build ./...` are clean.
- Pre-existing, unrelated issue noted but not fixed (out of scope for this diff): `npm run lint` fails because no ESLint config file (`.eslintrc*` / `eslint.config.*`) exists in the repo.

## Checklist

- [x] Tests pass — `go test ./...` (21/21 backend packages green)
- [x] Tests pass — `vitest run` (30/30 frontend files, 338 tests)
- [x] Coverage ≥90% for all packages achievable without live infrastructure (unchanged from last audit)
- [x] Breaking changes addressed (none applied; all documented in `.codekeep/deps.md`)
- [x] No push to main/master
