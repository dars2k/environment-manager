## Summary

- **P1 – Dependency updates**: Applied semver-safe minor/patch bumps for Go (16 packages: validator/v10, crypto, sonic, snappy, compress, cpuid/v2, go-isatty, stats, go-toml/v2, quic-go, mongo-driver/v2, arch, net, sync, sys, text) and npm (17 packages: @mui/x-data-grid, @tanstack/react-query, axios, react, react-dom, react-router-dom, @types/node, @types/react, @types/react-dom, @typescript-eslint/eslint-plugin, @typescript-eslint/parser, @vitest/coverage-v8, @vitest/ui, eslint, eslint-plugin-react-hooks, prettier, vitest).
- **P2 – Coverage uplift**: Confirmed overall backend coverage remains high at 92.3% statement coverage. Confirmed frontend statement coverage remains high at 92.31%.
- **P3 – QA**: Performed static analysis and verified frontend and backend test suites, including race condition detection.

## Dependency table

### Go (backend)

| Package | Old | New | Breaking? |
|---|---|---|---|
| github.com/go-playground/validator/v10 | v10.30.2 | v10.30.3 | No |
| golang.org/x/crypto | v0.52.0 | v0.54.0 | No |
| github.com/bytedance/sonic | v1.15.1 | v1.15.2 | No |
| github.com/golang/snappy | v0.0.4 | v1.0.0 | No |
| github.com/klauspost/compress | v1.18.6 | v1.19.0 | No |
| github.com/klauspost/cpuid/v2 | v2.3.0 | v2.4.0 | No |
| github.com/mattn/go-isatty | v0.0.22 | v0.0.23 | No |
| github.com/montanaflynn/stats | v0.9.0 | v0.12.2 | No |
| github.com/pelletier/go-toml/v2 | v2.3.1 | v2.4.3 | No |
| github.com/quic-go/quic-go | v0.59.1 | v0.60.0 | No |
| go.mongodb.org/mongo-driver/v2 | v2.6.0 | v2.8.0 | No |
| golang.org/x/arch | v0.26.0 | v0.29.0 | No |
| golang.org/x/net | v0.55.0 | v0.57.0 | No |
| golang.org/x/sync | v0.20.0 | v0.22.0 | No |
| golang.org/x/sys | v0.45.0 | v0.47.0 | No |
| golang.org/x/text | v0.37.0 | v0.40.0 | No |

### npm (frontend)

| Package | Old | New | Breaking? |
|---|---|---|---|
| @mui/x-data-grid | 6.18.3 | 6.20.4 | No |
| @tanstack/react-query | 5.100.8 | 5.101.2 | No |
| axios | 1.15.2 | 1.18.1 | No |
| react | 18.2.0 | 18.3.1 | No |
| react-dom | 18.2.0 | 18.3.1 | No |
| react-router-dom | 6.30.1 | 6.30.4 | No |
| @types/node | 22.19.15 | 22.20.1 | No |
| @types/react | 18.3.28 | 18.3.31 | No |
| @types/react-dom | 18.2.17 | 18.3.7 | No |
| @typescript-eslint/eslint-plugin | 6.13.2 | 6.21.0 | No |
| @typescript-eslint/parser | 6.13.2 | 6.21.0 | No |
| @vitest/coverage-v8 | 3.2.4 | 3.2.7 | No |
| @vitest/ui | 3.2.4 | 3.2.7 | No |
| eslint | 8.55.0 | 8.57.1 | No |
| eslint-plugin-react-hooks | 4.6.0 | 4.6.2 | No |
| prettier | 3.8.1 | 3.9.5 | No |
| vitest | 3.2.4 | 3.2.7 | No |

## Coverage: before → after

| Package | Before | After |
|---|---|---|
| Backend overall | 92.2% | **92.3%** ✅ |
| Frontend overall | 92.29% | **92.31%** ✅ |

## QA fixes list

1. Verified concurrency safety using backend Go race detector (`go test -race ./...`).
2. Confirmed that no warnings or deprecations are introduced with the minor/patch bumps.

## Checklist

- [x] tests pass
- [x] coverage ≥90%
- [x] breaking changes addressed
