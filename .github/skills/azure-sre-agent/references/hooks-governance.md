# Hooks and Governance Controls

Use this guide to enforce quality and safety gates in production.

## Hook Types

Two hook events are supported:

1. `Stop`: intercept final response before completion.
2. `PostToolUse`: inspect tool outcomes after execution.

Two configuration levels:

1. Agent-level hooks (apply broadly).
2. Custom-agent-level hooks (apply only to that custom agent).

## Prompt vs Command Hooks

1. Prompt hooks: nuanced LLM evaluation and policy checks.
2. Command hooks: deterministic checks (regex/policy/audit logic).

## Approval-Gate Pattern for Sensitive Actions

Use a `Stop` prompt hook for write-action gates.

Template:

```yaml
name: remediation-approval
event_type: Stop
activation_mode: always
description: Require explicit approval for write actions.
hook:
  type: prompt
  prompt: |
    Review the response for infrastructure-modifying actions.
    If any write/change action is present, reject and request explicit user approval.
    If read-only/reporting only, approve.
    $ARGUMENTS
    Return JSON:
    {"ok": true, "reason": "Read-only response"}
    or
    {"ok": false, "reason": "Approval required before remediation."}
  timeout: 30
  fail_mode: Block
  max_rejections: 3
```

## Audit Pattern for Tool Use

Use a `PostToolUse` command hook for audit trails.

Template:

```yaml
hooks:
  PostToolUse:
    - type: command
      matcher: "*"
      timeout: 30
      failMode: allow
      script: |
        #!/usr/bin/env python3
        import json, sys
        context = json.load(sys.stdin)
        tool = context.get("tool_name", "unknown")
        print(json.dumps({
          "decision": "allow",
          "hookSpecificOutput": {
            "additionalContext": f"[AUDIT] Tool executed: {tool}"
          }
        }))
```

## Safe Defaults

1. Always include `reason` when rejecting, especially for `Stop`.
2. Keep timeout short (`30s` default), increase only when needed.
3. Use `failMode: allow` for observability hooks and `failMode: block` for strict policy.
4. Set `maxRejections` to prevent endless loops (default 3, valid 1-25).
5. Log diagnostics to stderr for command hooks.

## API and UX Notes

1. Use v2 extended-agent APIs for full hook configuration.
2. Agent Canvas YAML view can omit hook details for v1 display.
3. Manage operationally in Builder > Hooks or v2 API workflows.

## AKS/Container Apps Governance Examples

Use approval gates for:

1. `az aks upgrade`, node pool scale, disruptive maintenance actions
2. Container App revision activation/deactivation or traffic shifts
3. NSG rule changes and subnet routing changes

Use audit hooks for:

1. all write-capable tool calls
2. high-risk command classes
3. privileged connector invocations

## KT Compliance Stop Hook (P1/P2 or Write Actions)

Use this pattern to enforce KT section completeness when required:

```yaml
name: kt-completeness-gate
event_type: Stop
activation_mode: always
description: Require KT sections for major incidents and production write actions.
hook:
  type: prompt
  prompt: |
    Review the response below.

    If incident severity is P1/P2 OR the response proposes a production write action,
    verify it includes these sections:
    1. Situation Appraisal
    2. Problem Analysis
    3. Decision Analysis
    4. Potential Problem Analysis

    If all required sections are present and meaningful, approve.
    Otherwise reject and explain which section is missing.

    $ARGUMENTS

    Return JSON:
    {"ok": true, "reason": "KT sections complete"}
    or
    {"ok": false, "reason": "Missing KT section(s): <list>."}
  timeout: 30
  fail_mode: Block
  max_rejections: 3
```

Pair with [kt-methodology.md](./kt-methodology.md) and
[kt-templates.md](./kt-templates.md).
Ready-to-use hook file: [../hooks/kt-completeness-gate.yaml](../hooks/kt-completeness-gate.yaml)
Bundle equivalent: [../bundles/governance-kt/hooks/kt-completeness-gate.yaml](../bundles/governance-kt/hooks/kt-completeness-gate.yaml)

## Sources

- Agent hooks overview: https://learn.microsoft.com/en-us/azure/sre-agent/agent-hooks
- Hook API tutorial: https://learn.microsoft.com/en-us/azure/sre-agent/tutorial-agent-hooks
- Official hook examples: https://github.com/microsoft/sre-agent/tree/main/samples/deployment-compliance/hooks
