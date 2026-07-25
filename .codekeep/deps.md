# Dependency Update Log

## Go (backend)

### Updates Applied (semver-safe) — 2026-07-25

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

Applied via `go get -u ./... && go mod tidy`; `go build ./...` and `go test ./... -race` pass. `github.com/golang/snappy` v1.0.0 (flagged as a major bump in the previous entry below) came along transitively this round with no code changes needed — the mongo-driver's use of it is unaffected.

### Major Bumps — Skipped (document only) — 2026-07-25

| Package | Current | Available | Reason Skipped |
|---|---|---|---|
| go.mongodb.org/mongo-driver (direct) | v1.17.9 | v2.x (`/v2` import path) | Deprecated upstream, but already the latest v1.x release — no semver-safe update available. Migrating to `/v2` changes import paths and several APIs (`mongo.Connect` signature, session handling) across ~15 files; needs a dedicated migration, not a dependency-bump pass. |

### Updates Applied (semver-safe) — 2026-05-23

| Package | Old | New | Breaking? |
|---|---|---|---|
| golang.org/x/crypto | v0.50.0 | v0.52.0 | No |
| golang.org/x/net | v0.53.0 | v0.55.0 | No |
| golang.org/x/sys | v0.43.0 | v0.45.0 | No |
| golang.org/x/text | v0.36.0 | v0.37.0 | No |
| github.com/cloudwego/base64x | v0.1.6 | v0.1.7 | No |
| github.com/montanaflynn/stats | v0.8.2 | v0.9.0 | No |
| github.com/quic-go/quic-go | v0.59.0 | v0.59.1 | No |

### Major Bumps — Skipped (document only)

| Package | Current | Available | Reason Skipped |
|---|---|---|---|
| github.com/golang/snappy | v0.0.4 | v1.0.0 | Major version bump; indirect dep; API may change |

## npm (frontend)

### Updates Applied (semver-safe) — 2026-07-25

| Package | Old | New | Breaking? |
|---|---|---|---|
| vite | 7.3.2 (exact pin) | 7.3.6 | No — patch-level, but pinned exactly so needed an explicit bump |
| esbuild, @typescript-eslint/*, jest-resolve-dependencies, @jest/reporters, react-router, react-router-dom (transitive) | — | in-range bump via `npm audit fix` | No |

`vite` was pinned to an exact version (no caret), so `npm audit fix` alone couldn't reach the patched 7.3.6 release that fixes two high-severity advisories (GHSA-v6wh-96g9-6wx3, GHSA-fx2h-pf6j-xcff). `npm audit fix` separately resolved a **critical** Vitest UI arbitrary-file-read issue (GHSA-5xrq-8626-4rwp) and a moderate `react-router`/`react-router-dom` open-redirect (GHSA-wrjc-x8rr-h8h6) that had in-range patched releases.

### Known residual advisory — 2026-07-25

- `esbuild` nested inside `vite@7.3.6` still reports a **low**-severity dev-server file-read issue (GHSA-g7r4-m6w7-qqqr, Windows-only, dev server only). No fix available without a `vite` major bump; low risk since this project's CI/prod doesn't run the Vite dev server on Windows.

### Pre-existing gap noticed during QA — 2026-07-25

`frontend/package.json` has ESLint as a dependency and an `npm run lint` script, but no `.eslintrc*` / `eslint.config.*` file exists anywhere in the repo — `npm run lint` fails immediately with "ESLint couldn't find a configuration file." This predates this change (confirmed via `git log`) and is unrelated to the dependency/coverage work here; flagging for a follow-up since it means lint has not actually been enforced.

### Updates Applied (semver-safe) — 2026-05-23

| Package | Old | New | Breaking? |
|---|---|---|---|
| @tanstack/react-query | 5.100.8 | 5.100.14 | No |
| axios | 1.15.2 | 1.14.0 | No (patch, latest 1.x available) |
| prettier | 3.8.1 | 3.8.3 | No |
| jest | 30.3.0 | 30.4.2 | No |
| jest-environment-jsdom | 30.3.0 | 30.4.1 | No |

### Major Bumps — Skipped (document only)

| Package | Current | Available | Reason Skipped |
|---|---|---|---|
| @mui/icons-material | 5.18.0 | 9.0.0 | Major; requires peer dep alignment |
| @mui/material | 5.18.0 | 9.0.0 | Major; breaking API changes |
| @mui/x-data-grid | 6.20.4 | 9.0.4 | Major; breaking API changes |
| @reduxjs/toolkit | 1.9.7 | 2.11.2 | Major; API changes |
| @types/react | 18.3.28 | 19.2.14 | Major; React 19 types |
| @types/react-dom | 18.3.7 | 19.2.3 | Major; React 19 types |
| @typescript-eslint/eslint-plugin | 6.21.0 | 8.59.1 | Major; requires ESLint 9 |
| @typescript-eslint/parser | 6.21.0 | 8.59.1 | Major; requires ESLint 9 |
| @vitejs/plugin-react | 4.7.0 | 6.0.1 | Major; requires Vite 8 |
| @vitest/coverage-v8 | 3.2.4 | 4.1.5 | Major; requires Vitest 4 |
| @vitest/ui | 3.2.4 | 4.1.5 | Major; requires Vitest 4 |
| date-fns | 2.30.0 | 4.1.0 | Major; breaking API changes |
| eslint | 8.57.1 | 10.3.0 | Major; flat config required |
| eslint-plugin-react-hooks | 4.6.2 | 7.1.1 | Major |
| jsdom | 26.1.0 | 29.1.1 | Major |
| react | 18.3.1 | 19.2.5 | Major; React 19 breaking changes |
| react-dom | 18.3.1 | 19.2.5 | Major; React 19 breaking changes |
| react-redux | 8.1.3 | 9.2.0 | Major |
| react-router-dom | 6.30.3 | 7.14.2 | Major; new data router API |
| recharts | 2.15.4 | 3.8.1 | Major |
| typescript | 5.9.3 | 6.0.3 | Major; strict type changes |
| vite | 7.3.2 | 8.0.10 | Major; config breaking changes |
| vitest | 3.2.4 | 4.1.5 | Major |
