# UI Confidence and Auditability

Frontend must make provenance explicit.

## Required UX Signals

- Badge for source:
  - `AI Generated` when `confidenceSource = ai_foundry`
  - `Rule Based` for heuristic/template outputs
- Tooltip/details showing:
  - confidence score
  - confidence source
  - correlation ID or run ID (if available)
- Audit trail section for AI-driven decisions:
  - input context summary
  - timestamp
  - model/provider metadata when available

## Minimal Rendering Pattern (React)

```tsx
<Badge appearance="tint" color={item.confidenceSource === 'ai_foundry' ? 'success' : 'warning'}>
  {item.confidenceSource === 'ai_foundry' ? 'AI Generated' : 'Rule Based'}
</Badge>
<Tooltip content={`Source: ${item.confidenceSource} | Score: ${item.confidenceScore ?? 'n/a'}`}>
  <span>Why this recommendation?</span>
</Tooltip>
```

## API Contract Reminder

Ensure these fields are present in API responses consumed by UI:

- `confidenceSource: string`
- `confidenceScore?: number`
- `correlationId?: string`
- `fallbackReason?: string`

