# Dependency Update Log

## Go (backend)

### Updates Applied (semver-safe) — 2026-07-18

| Package | Old | New | Breaking? |
|---|---|---|---|
| github.com/go-playground/validator/v10 | v10.30.2 | v10.30.3 | No |
| github.com/bytedance/sonic | v1.15.1 | v1.15.2 | No |
| github.com/golang/protobuf | v1.5.0 | v1.5.4 | No (deprecated module, patch only) |
| github.com/google/gofuzz | v1.0.0 | v1.2.0 | No |
| github.com/klauspost/compress | v1.18.6 | v1.19.0 | No |
| github.com/klauspost/cpuid/v2 | v2.3.0 | v2.4.0 | No |
| github.com/mattn/go-isatty | v0.0.22 | v0.0.23 | No |
| github.com/pelletier/go-toml/v2 | v2.3.1 | v2.4.3 | No |
| github.com/rogpeppe/go-internal | v1.10.0 | v1.15.0 | No (test tooling only) |
| github.com/yuin/goldmark | v1.4.13 | v1.8.4 | No |
| golang.org/x/arch | v0.26.0 | v0.29.0 | No |
| golang.org/x/crypto | v0.52.0 | v0.54.0 | No |
| golang.org/x/mod | v0.35.0 | v0.38.0 | No |
| golang.org/x/net | v0.55.0 | v0.57.0 | No |
| golang.org/x/sync | v0.20.0 | v0.22.0 | No |
| golang.org/x/sys | v0.45.0 | v0.47.0 | No |
| golang.org/x/term | v0.43.0 | v0.45.0 | No |
| golang.org/x/text | v0.37.0 | v0.40.0 | No |
| golang.org/x/tools | v0.44.0 | v0.48.0 | No (build tooling only) |

### Major Bumps — Skipped (document only)

| Package | Current | Available | Reason Skipped |
|---|---|---|---|
| github.com/golang/snappy | v0.0.4 | v1.0.0 | Major version bump; indirect dep; API may change |
| github.com/montanaflynn/stats | v0.9.0 | v0.12.2 | Indirect (pre-1.0), skipped to avoid unreviewed transitive behavior change |
| github.com/quic-go/quic-go | v0.59.1 | v0.60.0 | Indirect (pre-1.0), skipped to avoid unreviewed transitive behavior change |
| go.mongodb.org/mongo-driver/v2 | v2.6.0 | v2.8.0 | Indirect; app uses v1 driver directly, v2 pulled transitively — left as-is |

## npm (frontend)

### Updates Applied (semver-safe) — 2026-07-18

| Package | Old | New | Breaking? |
|---|---|---|---|
| @tanstack/react-query | 5.100.8 | 5.101.2 | No |

### Not updated — registry mirror has no newer version than installed

| Package | Installed | Registry `latest` | Note |
|---|---|---|---|
| axios | 1.15.2 | 1.14.0 | Registry mirror lags behind installed version; would be a downgrade, skipped |
| prettier | 3.8.1 | 3.8.1 | Registry mirror reports current version as latest |

### Major Bumps — Skipped (document only, unchanged from last audit)

| Package | Current | Available | Reason Skipped |
|---|---|---|---|
| @mui/icons-material | 5.18.0 | 9.2.0 | Major; requires peer dep alignment |
| @mui/material | 5.18.0 | 9.2.0 | Major; breaking API changes |
| @mui/x-data-grid | 6.18.3 | 9.10.0 | Major; breaking API changes |
| @reduxjs/toolkit | 1.9.7 | 2.12.0 | Major; API changes |
| @types/react | 18.3.28 | 19.2.17 | Major; React 19 types |
| @types/react-dom | 18.2.17 | 19.2.3 | Major; React 19 types |
| @typescript-eslint/eslint-plugin | 6.13.2 | 8.64.0 | Major; requires ESLint 9 |
| @typescript-eslint/parser | 6.13.2 | 8.64.0 | Major; requires ESLint 9 |
| @vitejs/plugin-react | 4.7.0 | 6.0.3 | Major; requires Vite 8 |
| @vitest/coverage-v8 | 3.2.4 | 4.1.10 | Major; requires Vitest 4 |
| @vitest/ui | 3.2.4 | 4.1.10 | Major; requires Vitest 4 |
| date-fns | 2.30.0 | 4.4.0 | Major; breaking API changes |
| eslint | 8.55.0 | 10.7.0 | Major; flat config required |
| eslint-plugin-react-hooks | 4.6.0 | 7.1.1 | Major |
| jsdom | 26.1.0 | 29.1.1 | Major |
| react | 18.2.0 | 19.2.7 | Major; React 19 breaking changes |
| react-dom | 18.2.0 | 19.2.7 | Major; React 19 breaking changes |
| react-redux | 8.1.3 | 9.3.0 | Major |
| react-router-dom | 6.30.1 | 7.18.1 | Major; new data router API |
| recharts | 2.15.4 | 3.9.2 | Major |
| typescript | 5.9.3 | 7.0.2 | Major; strict type changes |
| vite | 7.3.2 | 8.1.5 | Major; config breaking changes |
| vitest | 3.2.4 | 4.1.10 | Major |

---

## Prior audit — 2026-05-23

### Go

| Package | Old | New | Breaking? |
|---|---|---|---|
| golang.org/x/crypto | v0.50.0 | v0.52.0 | No |
| golang.org/x/net | v0.53.0 | v0.55.0 | No |
| golang.org/x/sys | v0.43.0 | v0.45.0 | No |
| golang.org/x/text | v0.36.0 | v0.37.0 | No |
| github.com/cloudwego/base64x | v0.1.6 | v0.1.7 | No |
| github.com/montanaflynn/stats | v0.8.2 | v0.9.0 | No |
| github.com/quic-go/quic-go | v0.59.0 | v0.59.1 | No |

### npm

| Package | Old | New | Breaking? |
|---|---|---|---|
| @tanstack/react-query | 5.100.8 | 5.100.14 | No |
| axios | 1.15.2 | 1.14.0 | No (patch, latest 1.x available) |
| prettier | 3.8.1 | 3.8.3 | No |
| jest | 30.3.0 | 30.4.2 | No |
| jest-environment-jsdom | 30.3.0 | 30.4.1 | No |
