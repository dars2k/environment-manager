# PR: chore(codekeep): dep updates + coverage + QA

## Deps table
| Package | Old | New | Breaking? |
|---------|-----|-----|-----------|
| github.com/golang/snappy | v0.0.4 | v1.0.0 | Yes (Major bump) |
| github.com/cloudwego/base64x | v0.1.6 | v0.1.7 | No |
| github.com/montanaflynn/stats | v0.8.2 | v0.9.0 | No |
| github.com/quic-go/quic-go | v0.59.0 | v0.59.1 | No |
| @tanstack/react-query | 5.100.8 | 5.100.14 | No |
| axios | 1.15.2 | 1.16.1 | No |
| jest | 30.3.0 | 30.4.2 | No |

## Coverage: before → after
- Backend: 91.9% → 95.4%
- Frontend: 92.3% → 92.6%

## QA fixes list
- Fixed goroutine leak in `websocket/client` by ensuring `Close()` closes the `send` channel and unregisters from hub.
- Added double-close protection in `websocket/client` using `recover`.
- Fixed unhandled rejection in `EditEnvironment` page by properly catching errors in `handleSubmit`.
- Removed accidental `frontend/bun.lock` file.

## Checklist
- [x] tests pass
- [x] coverage ≥90%
- [x] breaking changes addressed (snappy update verified via tests)
