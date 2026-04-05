# Sources (Verified 2026-04-03)

Primary Microsoft and official project sources used for this skill:

- Microsoft Learn overview (GA status, framework scope):  
  https://learn.microsoft.com/agent-framework/overview
- Microsoft Foundry Agent Service (managed/hosted agent service overview):  
  https://learn.microsoft.com/azure/ai-foundry/agents/overview
- Microsoft Foundry + Agent Framework integration guidance:  
  https://learn.microsoft.com/azure/ai-foundry/agents/agent-framework
- Workflow pattern docs (Sequential, Concurrent, Handoff, Group Chat, Magentic):  
  https://learn.microsoft.com/agent-framework/user-guide/workflows/
- C# middleware docs (`runFunc` + `runStreamingFunc`):  
  https://learn.microsoft.com/agent-framework/user-guide/agents/middleware/
- AG-UI integration docs (streaming, approvals, predictive updates, shared state):  
  https://learn.microsoft.com/agent-framework/user-guide/agents/agent-ui/
- MCP integration docs (local/hosted tools, safety/cost cautions):  
  https://learn.microsoft.com/agent-framework/user-guide/agents/model-context-protocol/
- A2A docs (interoperability and task model):  
  https://learn.microsoft.com/agent-framework/user-guide/agents/agent-to-agent/
- Known AG-UI approval/tool-call issue and fix context:  
  https://github.com/microsoft/agent-framework/issues/1850

### Added 2026-03-16 (Context7-verified)

- Durable Agents (Azure Functions hosting, HTTP endpoints, checkpoint/resume):  
  https://github.com/microsoft/agent-framework/blob/main/docs/features/durable-agents/README.md
- Durable Agents TTL configuration (global + per-agent):  
  https://github.com/microsoft/agent-framework/blob/main/docs/features/durable-agents/durable-agents-ttl.md
- Azure Functions hosting (`ConfigureDurableAgents`, auto HTTP endpoints):  
  https://github.com/microsoft/agent-framework/blob/main/dotnet/src/Microsoft.Agents.AI.Hosting.AzureFunctions/README.md
- Structured Output (`ResponseFormat`, `ForJsonSchema<T>`, inter-agent data flow):  
  https://github.com/microsoft/agent-framework/blob/main/docs/decisions/0016-structured-output.md
- Agent tools architecture (MCP tool filtering, OpenAI remote MCP servers):  
  https://github.com/microsoft/agent-framework/blob/main/docs/decisions/0002-agent-tools.md
- Long-running operations (cancel run, thread management):  
  https://github.com/microsoft/agent-framework/blob/main/docs/decisions/0009-support-long-running-operations.md
- Foundry SDK alignment (create/get/delete agents, thread usage, sequential orchestration):  
  https://github.com/microsoft/agent-framework/blob/main/docs/specs/001-foundry-sdk-alignment.md
- Declarative workflows (YAML-based action definitions):  
  https://github.com/microsoft/agent-framework/blob/main/python/samples/03-workflows/declarative/README.md
- OpenTelemetry observability (`configure_otel`, automatic agent/LLM/tool tracing):  
  Context7: /microsoft/agent-framework (`configure_otel` service_name + Azure Monitor integration)
- Human-in-the-loop tool approval (`approval_mode`, `always_require`, approval handlers):  
  Context7: /microsoft/agent-framework (HITL approval patterns)
- Workflow checkpointing (`FileCheckpointStorage`, resume from checkpoint):  
  Context7: /microsoft/agent-framework (checkpoint storage and resume)

When updating this skill, re-check these pages and adjust guidance if capability support has changed.
