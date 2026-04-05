# Run Chaos Experiment

## Steps

### Step 1 — Define the Hypothesis

Before running any experiment, document:

```yaml
experiment:
  name: <descriptive-name>
  hypothesis: "<When X fault occurs, the system should Y>"
  target: <resource-name>
  steady_state:
    - <metric-1 is within normal range>
    - <metric-2 is within normal range>
  abort_conditions:
    - <condition that triggers immediate stop>
```

### Step 2 — Enable Chaos Target

Ensure the target resource has Chaos Studio enabled:

```bash
# Check if target is enabled
az rest --method get \
  --url "https://management.azure.com/<resource-id>/providers/Microsoft.Chaos/targets?api-version=2025-01-01"

# Enable target (if not already)
az rest --method put \
  --url "https://management.azure.com/<resource-id>/providers/Microsoft.Chaos/targets/Microsoft-AppService?api-version=2025-01-01" \
  --body '{"properties": {}}'
```

### Step 3 — Verify Monitoring

Before injecting faults:

- Application Insights dashboard is open
- Alerts are active for key metrics
- Team is aware the experiment is running
- Abort procedure is documented

### Step 4 — Run the Experiment

```bash
# Start the experiment
az rest --method post \
  --url "https://management.azure.com/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Chaos/experiments/{name}/start?api-version=2025-01-01"

# Check status
az rest --method get \
  --url "https://management.azure.com/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Chaos/experiments/{name}/executions?api-version=2025-01-01"
```

### Step 5 — Observe and Record

During the experiment:

- Monitor steady-state metrics
- Record time to detection (when alerts fire)
- Record time to mitigation (when system recovers)
- Note any unexpected behaviour

If abort conditions are met:

```bash
# Cancel the experiment
az rest --method post \
  --url "https://management.azure.com/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Chaos/experiments/{name}/cancel?api-version=2025-01-01"
```

### Step 6 — Review and Act

After the experiment:

- Compare actual behaviour against hypothesis
- Document findings and gaps
- Create action items for resilience improvements
- Update runbooks if recovery procedures were inadequate
- Schedule follow-up experiment after fixes are implemented
