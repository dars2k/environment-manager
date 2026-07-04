# Dependency Update Log

## Go (backend)

### Updates Applied (semver-safe) — 2026-07-04

Ran `go get -u ./...` + `go mod tidy`. All are patch/minor bumps of transitive dependencies; `go build ./...` and `go test ./...` pass.

| Package | Old | New | Breaking? |
|---|---|---|---|
| github.com/bytedance/sonic | v1.15.1 | v1.15.2 | No |
| github.com/go-playground/validator/v10 | v10.30.2 | v10.30.3 | No |
| github.com/klauspost/compress | v1.18.6 | v1.19.0 | No |
| github.com/klauspost/cpuid/v2 | v2.3.0 | v2.4.0 | No |
| github.com/pelletier/go-toml/v2 | v2.3.1 | v2.4.2 | No |
| github.com/quic-go/quic-go | v0.59.1 | v0.60.0 | No |
| go.mongodb.org/mongo-driver/v2 | v2.6.0 | v2.7.0 | No |
| golang.org/x/arch | v0.26.0 | v0.28.0 | No |
| golang.org/x/crypto | v0.52.0 | v0.53.0 | No |
| golang.org/x/net | v0.55.0 | v0.56.0 | No |
| golang.org/x/sync | v0.20.0 | v0.21.0 | No |
| golang.org/x/sys | v0.45.0 | v0.46.0 | No |
| golang.org/x/text | v0.37.0 | v0.38.0 | No |

### Notable transitive bump (documented, not a manual choice)

| Package | Old | New | Note |
|---|---|---|---|
| github.com/golang/snappy | v0.0.4 | v1.0.0 | Pulled in transitively by `go get -u`. Crosses the 0.x→1.x boundary but same import path (no `/v2`); not imported directly by our code (only via mongo-driver's wire compression). `go build`/`go test` pass. |

### Major Bumps — Skipped (document only)

No direct major bumps were applied. `go list -u -m all` shows majors available for a couple of indirect/tooling deps (e.g. `github.com/rogpeppe/go-internal`, `github.com/yuin/goldmark`, `golang.org/x/tools`) that are test/build tooling only, pulled transitively — left as-is.

## npm (frontend)

### Updates Applied (semver-safe) — 2026-07-04

| Package | Old | New | Breaking? |
|---|---|---|---|
| vitest | ^3.2.4 | ^3.2.6 | No — fixes critical CVE (GHSA-5xrq-8626-4rwp, arbitrary file read/execute via Vitest UI server) |
| @vitest/coverage-v8 | ^3.2.4 | ^3.2.6 | No |
| @vitest/ui | ^3.2.4 | ^3.2.6 | No |
| vite | ^7.3.2 (pinned 7.3.2) | ^7.3.6 | No — fixes high CVEs (GHSA-v6wh-96g9-6wx3, GHSA-fx2h-pf6j-xcff) |
| react-router-dom | ^6.30.1 | ^6.30.4 | No — fixes moderate open-redirect CVE (GHSA-2j2x-hqr9-3h42) |
| @tanstack/react-query | ^5.100.8 | ^5.101.2 | No |
| @mui/x-data-grid | ^6.18.3 | resolves to 6.20.4 in-range | No |

### Transitive CVE fixes via `overrides` (following the existing `minimatch` override pattern)

| Package | Old (resolved) | New (override) | Advisory |
|---|---|---|---|
| ws | 8.17.1 / 8.18.2 | ^8.18.3 | GHSA-58qx-3vcg-4xpx, GHSA-96hv-2xvq-fx4p |
| js-yaml | 4.1.1 | ^4.2.0 | GHSA-h67p-54hq-rp68 |
| ajv | 6.12.6 | ^6.14.0 | GHSA-2g4f-4pwh-qvx6 |
| brace-expansion | 2.0.2 | ^2.0.3 | GHSA-f886-m6hf-6m8v |
| @babel/core | 7.29.0 | ^7.29.7 | GHSA-4x5r-pxfx-6jf8 |
| esbuild | 0.27.7 | ^0.28.1 | GHSA-g7r4-m6w7-qqqr |
| engine.io-client | 6.6.3 | ^6.6.6 | consumes the `ws` fix above |
| yaml | 1.10.2 | ^2.8.2 | GHSA-48c2-rrv3-qjmp (quadratic-complexity DoS) |

Verified after applying: `npx tsc --noEmit`, `npm run build`, and the full `vitest run` suite all pass with these overrides in place.

### Not fixed — environment/registry limitation

`npm audit` also flags `axios` (high, several advisories) and its transitive `form-data` (high, GHSA-hmw2-7cc7-3qxx). The registry reachable from this environment does not serve any `axios` version newer than 1.14.0 (`npm view axios versions`/`dist-tags` cap out there), which is *older* than the 1.15.2 already pinned in `package.json` — `npm install axios@<anything newer>` fails with `ETARGET` for every version above 1.14.0. This looks like a stale/frozen registry mirror in this sandbox rather than a real absence of fixed releases. Recommend re-running `npm audit fix` for axios/form-data from an environment with full npm registry access.

### Major Bumps — Skipped (document only)

| Package | Current | Available | Reason Skipped |
|---|---|---|---|
| @mui/icons-material / @mui/material | 5.18.0 | 9.x | Major; breaking API changes |
| @reduxjs/toolkit | 1.9.7 | 2.x | Major; API changes |
| react / react-dom | 18.3.1 | 19.x | Major; React 19 breaking changes |
| react-redux | 8.1.3 | 9.x | Major |
| react-router-dom | 6.30.4 | 7.x | Major; new data router API |
| recharts | 2.15.4 | 3.x | Major |
| date-fns | 2.30.0 | 4.x | Major; breaking API changes |
| eslint | 8.57.1 | 10.x | Major; flat config required |
| eslint-plugin-react-hooks | 4.6.2 | 7.x | Major |
| @typescript-eslint/eslint-plugin / parser | 6.21.0 | 8.x | Major; requires ESLint 9 |
| @vitejs/plugin-react | 4.7.0 | 6.x | Major; requires Vite 8 |
| typescript | 5.9.3 | 6.x | Major |
| jsdom | 26.1.0 | 29.x | Major |
| @types/react / @types/react-dom | 18.3.x | 19.x | Major; React 19 types |
