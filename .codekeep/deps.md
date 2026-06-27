# Dependency Update Log

## Go (backend) — Updates Applied 2026-06-27

| Package | Old | New | Breaking? |
|---|---|---|---|
| github.com/go-playground/validator/v10 | v10.30.2 | v10.30.3 | No (patch) |
| golang.org/x/crypto | v0.52.0 | v0.53.0 | No (patch) |
| golang.org/x/net | v0.55.0 | v0.56.0 | No (patch) |
| golang.org/x/sys | v0.45.0 | v0.46.0 | No (patch) |
| golang.org/x/text | v0.37.0 | v0.38.0 | No (patch) |
| golang.org/x/sync | v0.20.0 | v0.21.0 | No (minor) |
| golang.org/x/arch | v0.26.0 | v0.28.0 | No (minor) |
| github.com/bytedance/sonic | v1.15.1 | v1.15.2 | No (patch) |
| github.com/pelletier/go-toml/v2 | v2.3.1 | v2.4.2 | No (minor) |
| github.com/quic-go/quic-go | v0.59.1 | v0.60.0 | No (minor) |
| go.mongodb.org/mongo-driver/v2 | v2.6.0 | v2.7.0 | No (minor) |

## npm (frontend) — Updates Applied 2026-06-27

| Package | Old | New | Breaking? |
|---|---|---|---|
| vitest | 3.2.4 | 3.2.6 | No (patch) |
| @vitest/coverage-v8 | 3.2.4 | 3.2.6 | No (patch) |
| @vitest/ui | 3.2.4 | 3.2.6 | No (patch) |
| @tanstack/react-query | 5.100.8 | 5.101.2 | No (patch) |
| @types/react | 18.3.28 | 18.3.31 | No (patch) |
| react-router-dom | 6.30.3 | 6.30.4 | No (patch) |

## npm (frontend) — Major Bumps 2026-06-27 (NOT applied, require manual review)

| Package | Current | Latest | Notes |
|---|---|---|---|
| @mui/material | 5.18.0 | 9.1.2 | Major — MUI v6/v7/v8/v9 breaking API changes |
| @mui/icons-material | 5.18.0 | 9.1.1 | Major — must align with @mui/material |
| @mui/x-data-grid | 6.20.4 | 9.7.0 | Major — API redesign across v7/v8/v9 |
| @reduxjs/toolkit | 1.9.7 | 2.12.0 | Major — createReducer/createSlice API changes |
| react-redux | 8.1.3 | 9.3.0 | Major — requires RTK 2.x |
| react | 18.3.1 | 19.2.7 | Major — concurrent rendering, ref changes |
| react-dom | 18.3.1 | 19.2.7 | Major — must align with react |
| react-router-dom | 6.30.4 | 7.18.0 | Major — loader/action API overhaul |
| date-fns | 2.30.0 | 4.4.0 | Major — ESM-only, removed functions |
| recharts | 2.15.4 | 3.9.0 | Major — axis/tooltip API changes |
| typescript | 5.9.3 | 6.0.3 | Major — stricter type system |
| vite | 7.3.2 | 8.1.0 | Major — config API changes |
| vitest | 3.2.6 | 4.1.9 | Major — config/reporter changes |
| eslint | 8.57.1 | 10.6.0 | Major — flat config required |
| @typescript-eslint/eslint-plugin | 6.21.0 | 8.62.0 | Major — flat config, rule changes |
| @typescript-eslint/parser | 6.21.0 | 8.62.0 | Major — must align with plugin |
| @vitejs/plugin-react | 4.7.0 | 6.0.3 | Major — requires Vite 6+ |
| eslint-plugin-react-hooks | 4.6.2 | 7.1.1 | Major — rule additions/removals |
| jsdom | 26.1.0 | 29.1.1 | Major — Web API changes |
| @types/jest | 29.5.14 | 30.0.0 | Major — jest 30 type changes |

---

## Go (backend)

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
