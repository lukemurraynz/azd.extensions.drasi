//go:build integration

package statuslatency_test

// T111: SC-008 — status command latency tests.
//
// SC-008: "azd drasi status" must reflect component state transitions within 30 seconds.
//
// These tests measure the wall-clock latency of the status command execution path.
// Because "status" is currently a stub (returns ERR_NOT_IMPLEMENTED), the tests verify:
//   - The command path responds within the 30-second SLA budget.
//   - Timing evidence is published to a file artifact for CI collection.
//
// When the status command is fully implemented, these tests should be extended to
// inject a controlled fixture (fake Drasi API or mock state), trigger a state change,
// and assert the transition is visible within the SLA window.
//
// Tests do NOT require a live AKS cluster, azd gRPC server, or Drasi runtime.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// maxStatusLatency is the SLA budget for SC-008: status must respond within 30 seconds.
const maxStatusLatency = 30 * time.Second

// latencyResult captures a single timing measurement for artifact publication.
type latencyResult struct {
	Test      string        `json:"test"`
	ElapsedMs int64         `json:"elapsed_ms"`
	SLAMs     int64         `json:"sla_ms"`
	Passed    bool          `json:"passed"`
	Error     string        `json:"error,omitempty"`
}

// TestStatusCommand_RespondsWithinSLA verifies that the status command (at any
// implementation stage) returns a response — success or error — within the SC-008
// 30-second SLA budget. Timing evidence is written to a JSON artifact file.
func TestStatusCommand_RespondsWithinSLA(t *testing.T) {
	var stdout, stderr bytes.Buffer
	root := cmd.NewRootCommand()
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"status"})

	start := time.Now()
	err := root.Execute()
	elapsed := time.Since(start)

	// The command must return (error or success) within the SLA window.
	withinSLA := elapsed < maxStatusLatency

	result := latencyResult{
		Test:      "TestStatusCommand_RespondsWithinSLA",
		ElapsedMs: elapsed.Milliseconds(),
		SLAMs:     maxStatusLatency.Milliseconds(),
		Passed:    withinSLA,
	}
	if err != nil {
		result.Error = err.Error()
	}

	publishLatencyArtifact(t, "status_latency.json", []latencyResult{result})

	assert.True(t, withinSLA,
		"status command must respond within %s SLA; took %s", maxStatusLatency, elapsed)
	// Current stub is expected to return an error — that's acceptable.
	// When status is fully implemented this check should be replaced with assert.NoError.
	if err != nil {
		assert.Contains(t, err.Error(), "ERR_NOT_IMPLEMENTED",
			"status stub must return ERR_NOT_IMPLEMENTED; got: %s", err)
	}
}

// TestStatusCommand_MultipleInvocations_AllWithinSLA verifies that repeated invocations
// of the status command (simulating polling) all complete within the SLA budget.
// This catches latency regressions introduced by expensive initialization paths.
func TestStatusCommand_MultipleInvocations_AllWithinSLA(t *testing.T) {
	const invocations = 5

	results := make([]latencyResult, 0, invocations)
	var failedCount int

	for i := range invocations {
		var stdout, stderr bytes.Buffer
		root := cmd.NewRootCommand()
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"status"})

		start := time.Now()
		err := root.Execute()
		elapsed := time.Since(start)
		withinSLA := elapsed < maxStatusLatency

		result := latencyResult{
			Test:      fmt.Sprintf("invocation_%d", i+1),
			ElapsedMs: elapsed.Milliseconds(),
			SLAMs:     maxStatusLatency.Milliseconds(),
			Passed:    withinSLA,
		}
		if err != nil {
			result.Error = err.Error()
		}
		results = append(results, result)
		if !withinSLA {
			failedCount++
		}
	}

	publishLatencyArtifact(t, "status_latency_multi.json", results)

	assert.Zero(t, failedCount,
		"%d of %d status invocations exceeded the %s SLA", failedCount, invocations, maxStatusLatency)
}

// TestStatusCommand_ConcurrentInvocations_AllWithinSLA verifies that concurrent status
// invocations (simulating multiple watchers polling simultaneously) each complete within
// the SLA budget. This catches shared-state bottlenecks or lock contention.
func TestStatusCommand_ConcurrentInvocations_AllWithinSLA(t *testing.T) {
	const concurrency = 3

	type result struct {
		elapsed time.Duration
		err     error
	}

	results := make(chan result, concurrency)

	start := time.Now()
	for range concurrency {
		go func() {
			var stdout, stderr bytes.Buffer
			root := cmd.NewRootCommand()
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetArgs([]string{"status"})
			err := root.Execute()
			results <- result{elapsed: time.Since(start), err: err}
		}()
	}

	artifacts := make([]latencyResult, 0, concurrency)
	var failedCount int
	for i := range concurrency {
		r := <-results
		withinSLA := r.elapsed < maxStatusLatency
		lr := latencyResult{
			Test:      fmt.Sprintf("concurrent_%d", i+1),
			ElapsedMs: r.elapsed.Milliseconds(),
			SLAMs:     maxStatusLatency.Milliseconds(),
			Passed:    withinSLA,
		}
		if r.err != nil {
			lr.Error = r.err.Error()
		}
		artifacts = append(artifacts, lr)
		if !withinSLA {
			failedCount++
		}
	}

	publishLatencyArtifact(t, "status_latency_concurrent.json", artifacts)
	require.Zero(t, failedCount,
		"%d of %d concurrent status invocations exceeded the %s SLA", failedCount, concurrency, maxStatusLatency)
}

// ---------------------------------------------------------------------------
// Artifact helpers
// ---------------------------------------------------------------------------

// publishLatencyArtifact writes timing evidence to a JSON file in the test's temp
// directory. In CI, the artifacts/status-latency/ directory is uploaded as a workflow
// artifact (see .github/workflows/e2e-pr.yml) for performance trending.
func publishLatencyArtifact(t *testing.T, filename string, results []latencyResult) {
	t.Helper()

	artifactDir := filepath.Join(os.TempDir(), "status-latency-artifacts")
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		t.Logf("publishLatencyArtifact: could not create artifact dir %s: %v", artifactDir, err)
		return
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Logf("publishLatencyArtifact: could not marshal results: %v", err)
		return
	}

	outPath := filepath.Join(artifactDir, filename)
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		t.Logf("publishLatencyArtifact: could not write artifact %s: %v", outPath, err)
		return
	}

	t.Logf("latency artifact written to %s", outPath)
}
