# Dependency Update Log

## Go (backend)

### Updates Applied (semver-safe)

| Package | Old | New | Breaking? |
|---|---|---|---|
| golang.org/x/crypto | v0.49.0 | v0.50.0 | No |
| golang.org/x/net | v0.52.0 | v0.53.0 | No |
| golang.org/x/sys | v0.42.0 | v0.43.0 | No |
| golang.org/x/text | v0.35.0 | v0.36.0 | No |
| golang.org/x/arch | v0.25.0 | v0.26.0 | No |
| golang.org/x/tools | v0.42.0 | v0.44.0 | No |
| go.mongodb.org/mongo-driver/v2 | v2.5.0 | v2.6.0 | No |
| github.com/bytedance/sonic | v1.15.0 | v1.15.1 | No |
| github.com/bytedance/sonic/loader | v0.5.0 | v0.5.1 | No |
| github.com/gin-contrib/sse | v1.1.0 | v1.1.1 | No |
| github.com/mattn/go-isatty | v0.0.20 | v0.0.22 | No |
| github.com/pelletier/go-toml/v2 | v2.2.4 | v2.3.1 | No |

### Major Bumps — Skipped (document only)

| Package | Current | Available | Reason Skipped |
|---|---|---|---|
| github.com/golang/snappy | v0.0.4 | v1.0.0 | Major version bump; indirect dep; API may change |

## npm (frontend)

### Updates Applied (semver-safe)

| Package | Old | New | Breaking? |
|---|---|---|---|
| @emotion/styled | 11.14.0 | 11.14.1 | No |
| @tanstack/react-query | 5.91.3 | 5.100.8 | No |
| axios | 1.15.0 | 1.14.0 | No (downgrade to latest available) |

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
