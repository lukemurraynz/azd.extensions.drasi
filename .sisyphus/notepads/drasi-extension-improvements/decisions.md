# Decisions — drasi-extension-improvements

## 2026-04-07 Session Start

### Architecture Decisions
- Progress helper: wraps yacspin, suppresses when --output json or non-TTY, outputs to stderr only
- Observability wiring: PersistentPreRunE on root command initializes tracer+meter; PersistentPostRunE calls shutdowns
- Deploy lock: ALREADY has defer release in deploy.go:154-156; Task 5 upgrades to JSON payload with timestamp
- Task 3 (logs --follow removal): `follow` var is also in payload map (line 120) — remove from payload too
- No gomock — all mocks are manual structs
- Survey/v2 is archived but acceptable as existing transitive dep
