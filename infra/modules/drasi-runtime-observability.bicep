targetScope = 'resourceGroup'

@description('OTLP endpoint URL to which Drasi runtime workloads export OpenTelemetry traces and metrics.')
param otlpEndpoint string

@description('Namespace containing the Drasi Kubernetes workloads.')
param drasiNamespace string = 'drasi-system'

// NOTE: The Go provision command consumes these outputs and patches the Drasi
// runtime workload specs after Azure resources are deployed. The probe values
// are configuration strings, not Kubernetes resources — the patch is applied
// via kubectl or the Drasi operator API, mirroring the workload-identity pattern.

output otlpExporterEndpoint string = otlpEndpoint
output drasiNamespace string = drasiNamespace

// livenessProbe: restarts the container when the Drasi runtime process deadlocks
// or becomes unresponsive to its internal health endpoint.
output livenessProbe string = 'httpGet: { path: /healthz/live, port: 8080 }, initialDelaySeconds: 15, periodSeconds: 20'

// readinessProbe: removes the pod from the Service endpoint slice until the
// runtime has finished initialising (connected to sources, registered reactions).
output readinessProbe string = 'httpGet: { path: /healthz/ready, port: 8080 }, initialDelaySeconds: 5, periodSeconds: 10'

// startupProbe: gives slow-starting containers up to 5 minutes before liveness
// checks begin, preventing premature restarts on cold-boot initialisation.
output startupProbe string = 'httpGet: { path: /healthz/startup, port: 8080 }, failureThreshold: 30, periodSeconds: 10'
