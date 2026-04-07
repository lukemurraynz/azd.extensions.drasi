# Draft: Drasi Azure Developer CLI Extension Improvements

## Requirements (confirmed)
- Plan for ALL 17 identified improvements
- User wants comprehensive work plan covering entire scope
- This is a Go extension for Azure Developer CLI
- Extension ID: `azure.drasi`, namespace: `drasi`

## Technical Decisions
- Language: Go (existing codebase)
- Test framework: Go standard testing + testify (from go.mod)
- CI: 80% coverage gate, race detector, golangci-lint
- azd SDK: azdext gRPC SDK
- Drasi CLI wrapper: internal/drasi/client.go

## Scope Boundaries
- INCLUDE: All 17 improvements from analysis
- INCLUDE: Tests for every change (maintain 80% coverage gate)
- EXCLUDE: New Drasi CLI features (this extension wraps existing CLI)
- EXCLUDE: Infrastructure/Bicep changes (unless needed for templates)

## Improvement Inventory (17 items)

### Critical — Functional Gaps
1. Diagnose stub checks (Key Vault, Log Analytics)
2. Deploy rollback on failure
3. Crash-safe deploy lock
4. Logs for all component kinds

### High — Developer Experience
5. Describe command
6. Status show all kinds by default
7. Progress indicators
8. Lifecycle hook implementations
9. Remove --follow no-op flag

### Medium — Code Quality
10. Externalize NetworkPolicy YAML
11. Wire observability package
12. Remove applyDefaultProviders no-op
13. Environment overlay example in sample

### Low — Feature Expansion
14. Additional scaffold templates (PostgreSQL, SQL Server, Kafka)
15. Interactive confirmations for destructive ops
16. Telemetry/usage tracking
17. Key Vault translator E2E validation tests

## Test Strategy Decision
- **Infrastructure exists**: YES (Go test + testify)
- **Automated tests**: YES (tests-after — maintain 80% coverage)
- **Framework**: Go standard testing + testify
- **Agent-Executed QA**: ALWAYS

## Open Questions
- None — user confirmed "plan for all"

## Research In Progress
- 3 explore agents gathering implementation details from all affected files
