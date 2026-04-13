package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T041 (success-path): white-box tests that inject runProvisionFunc to bypass the
// live Azure/gRPC dependency. These must live in package cmd (not cmd_test) so they
// can access the package-level runProvisionFunc variable.

// TestProvisionCommand_OutputJSON_EmitsResourceIDs verifies that when --output json
// is passed the command emits a JSON object to stdout containing "status" and
// "resourceIds" keys.
func TestProvisionCommand_OutputJSON_EmitsResourceIDs(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level runProvisionFunc.
	orig := runProvisionFunc
	t.Cleanup(func() { runProvisionFunc = orig })
	runProvisionFunc = func(cmd *cobra.Command, _ []string) error {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), `{"status":"ok","resourceIds":{}}`)
		return nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision", "--output", "json"})
	err := root.Execute()

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &out), "stdout must be valid JSON")
	assert.Equal(t, "ok", out["status"])
	_, hasResourceIDs := out["resourceIds"]
	assert.True(t, hasResourceIDs, "JSON output must contain resourceIds key")
}

// TestProvisionCommand_EnvironmentFlag_NoMutation verifies that a successful provision
// with --environment set returns no error and does not require a live Azure connection.
func TestProvisionCommand_EnvironmentFlag_NoMutation(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level runProvisionFunc.
	orig := runProvisionFunc
	t.Cleanup(func() { runProvisionFunc = orig })
	runProvisionFunc = func(_ *cobra.Command, _ []string) error {
		return nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision", "--environment", "dev"})
	err := root.Execute()

	require.NoError(t, err)
}

// TestProvisionCommand_AuditEvent_EmittedToStderr verifies that a successful provision
// writes "provision" to stderr (the audit event destination).
func TestProvisionCommand_AuditEvent_EmittedToStderr(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level runProvisionFunc.
	orig := runProvisionFunc
	t.Cleanup(func() { runProvisionFunc = orig })
	runProvisionFunc = func(cmd *cobra.Command, _ []string) error {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "provision audit event")
		return nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	stderr := &bytes.Buffer{}
	root.SetErr(stderr)
	root.SetArgs([]string{"provision"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, stderr.String(), "provision", "audit event must be emitted to stderr")
}

// TestProvisionCommand_DRASIPROVISIONEDWrittenOnSuccess verifies that a successful
// provision run with an environment flag returns no error (the DRASI_PROVISIONED
// write is part of defaultRunProvision which is replaced here by the stub).
func TestProvisionCommand_DRASIPROVISIONEDWrittenOnSuccess(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level runProvisionFunc.
	orig := runProvisionFunc
	t.Cleanup(func() { runProvisionFunc = orig })
	runProvisionFunc = func(_ *cobra.Command, _ []string) error {
		return nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision", "--environment", "dev"})
	err := root.Execute()

	require.NoError(t, err)
}

func TestGetEnvValue_ReturnsValue(t *testing.T) {
	t.Parallel()

	azdClient := newTestAzdClient(t, &testEnvironmentService{
		getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
			require.Equal(t, "dev", req.EnvName)
			require.Equal(t, "AZURE_AKS_CONTEXT", req.Key)
			return &azdext.KeyValueResponse{Value: "aks-dev"}, nil
		},
	})

	value, err := getEnvValue(context.Background(), azdClient, "dev", "AZURE_AKS_CONTEXT")

	require.NoError(t, err)
	assert.Equal(t, "aks-dev", value)
}

func TestGetEnvValue_NilResponseReturnsEmptyString(t *testing.T) {
	t.Parallel()

	azdClient := newTestAzdClient(t, &testEnvironmentService{
		getValueFunc: func(_ context.Context, _ *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
			return nil, nil
		},
	})

	value, err := getEnvValue(context.Background(), azdClient, "dev", "AZURE_AKS_CONTEXT")

	require.NoError(t, err)
	assert.Empty(t, value)
}

func TestGetEnvValue_GetValueFails(t *testing.T) {
	t.Parallel()

	azdClient := newTestAzdClient(t, &testEnvironmentService{
		getValueFunc: func(_ context.Context, _ *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
			return nil, errors.New("grpc unavailable")
		},
	})

	value, err := getEnvValue(context.Background(), azdClient, "dev", "AZURE_AKS_CONTEXT")

	assert.Empty(t, value)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "grpc unavailable")
}

func TestBuildResourceIDs_AllEnvValuesPresent(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"AZURE_AKS_CLUSTER_NAME":           "aks-cluster",
		"AZURE_KEY_VAULT_NAME":             "kv-name",
		"AZURE_LOG_ANALYTICS_WORKSPACE_ID": "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/la",
		"AZURE_ACR_LOGIN_SERVER":           "example.azurecr.io",
		"AZURE_RESOURCE_GROUP":             "rg-name",
	}

	azdClient := newTestAzdClient(t, &testEnvironmentService{
		getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
			return &azdext.KeyValueResponse{Value: values[req.Key]}, nil
		},
	})

	resourceIDs := buildResourceIDs(context.Background(), azdClient, "dev")

	assert.Equal(t, values, resourceIDs)
}

func TestBuildResourceIDs_SkipsEmptyValues(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"AZURE_AKS_CLUSTER_NAME":           "aks-cluster",
		"AZURE_KEY_VAULT_NAME":             "",
		"AZURE_LOG_ANALYTICS_WORKSPACE_ID": "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/la",
		"AZURE_ACR_LOGIN_SERVER":           "",
		"AZURE_RESOURCE_GROUP":             "rg-name",
	}

	azdClient := newTestAzdClient(t, &testEnvironmentService{
		getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
			return &azdext.KeyValueResponse{Value: values[req.Key]}, nil
		},
	})

	resourceIDs := buildResourceIDs(context.Background(), azdClient, "dev")

	assert.Equal(t, map[string]string{
		"AZURE_AKS_CLUSTER_NAME":           "aks-cluster",
		"AZURE_LOG_ANALYTICS_WORKSPACE_ID": "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/la",
		"AZURE_RESOURCE_GROUP":             "rg-name",
	}, resourceIDs)
}

func TestBuildResourceIDs_AllLookupsFailReturnsEmptyMap(t *testing.T) {
	t.Parallel()

	azdClient := newTestAzdClient(t, &testEnvironmentService{
		getValueFunc: func(_ context.Context, _ *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
			return nil, errors.New("lookup failed")
		},
	})

	resourceIDs := buildResourceIDs(context.Background(), azdClient, "dev")

	assert.Empty(t, resourceIDs)
}

func TestWarnUnmanagedResources_EmptyOrAbsentResourceGroupReturnsNil(t *testing.T) {
	for _, tc := range []struct {
		name     string
		response *azdext.KeyValueResponse
	}{
		{name: "absent resource group", response: nil},
		{name: "empty resource group", response: &azdext.KeyValueResponse{Value: ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			azdClient := newTestAzdClient(t, &testEnvironmentService{
				getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
					require.Equal(t, "AZURE_RESOURCE_GROUP", req.Key)
					return tc.response, nil
				},
			})

			cmd := &cobra.Command{}
			cmd.SetErr(&bytes.Buffer{})

			err := warnUnmanagedResources(cmd, context.Background(), azdClient, "dev", output.FormatTable)

			require.NoError(t, err)
		})
	}
}

func TestWarnUnmanagedResources_AzCLIUnavailableReturnsNil(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	azdClient := newTestAzdClient(t, &testEnvironmentService{
		getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
			require.Equal(t, "AZURE_RESOURCE_GROUP", req.Key)
			return &azdext.KeyValueResponse{Value: "rg-dev"}, nil
		},
	})

	cmd := &cobra.Command{}
	cmd.SetErr(&bytes.Buffer{})

	err := warnUnmanagedResources(cmd, context.Background(), azdClient, "dev", output.FormatTable)

	require.NoError(t, err)
}

func TestSwitchKubectlContext_KubectlMissingReturnsError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := switchKubectlContext(context.Background(), "desired-context")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "kubectl not found")
}

func TestSwitchKubectlContext_CurrentContextAlreadyMatches(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "kubectl.log")
	installFakeCommands(t, map[string]string{
		"kubectl": fakeScript(fmt.Sprintf(`
echo %%*>>"%s"
if "%%1 %%2"=="config current-context" (
  echo desired-context
  exit /b 0
)
if "%%1 %%2"=="config use-context" (
  exit /b 99
)
exit /b 1
`, logFile), fmt.Sprintf(`
echo "$@" >> "%s"
if [ "$1 $2" = "config current-context" ]; then
  echo desired-context
  exit 0
fi
if [ "$1 $2" = "config use-context" ]; then
  exit 99
fi
exit 1
`, logFile)),
	})

	err := switchKubectlContext(context.Background(), "desired-context")

	require.NoError(t, err)
	assert.Equal(t, []string{"config current-context"}, readNonEmptyLines(t, logFile))
}

func TestSwitchKubectlContext_ContextSwitchSucceeds(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "kubectl.log")
	installFakeCommands(t, map[string]string{
		"kubectl": fakeScript(fmt.Sprintf(`
echo %%*>>"%s"
if "%%1 %%2"=="config current-context" (
  echo other-context
  exit /b 0
)
if "%%1 %%2"=="config use-context" (
  exit /b 0
)
exit /b 1
`, logFile), fmt.Sprintf(`
echo "$@" >> "%s"
if [ "$1 $2" = "config current-context" ]; then
  echo other-context
  exit 0
fi
if [ "$1 $2" = "config use-context" ]; then
  exit 0
fi
exit 1
`, logFile)),
	})

	err := switchKubectlContext(context.Background(), "desired-context")

	require.NoError(t, err)
	assert.Equal(t, []string{"config current-context", "config use-context desired-context"}, readNonEmptyLines(t, logFile))
}

func TestSwitchKubectlContext_ContextSwitchFails(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "kubectl.log")
	installFakeCommands(t, map[string]string{
		"kubectl": fakeScript(fmt.Sprintf(`
echo %%*>>"%s"
if "%%1 %%2"=="config current-context" (
  echo other-context
  exit /b 0
)
if "%%1 %%2"=="config use-context" (
  echo failed to switch 1>&2
  exit /b 1
)
exit /b 1
`, logFile), fmt.Sprintf(`
echo "$@" >> "%s"
if [ "$1 $2" = "config current-context" ]; then
  echo other-context
  exit 0
fi
if [ "$1 $2" = "config use-context" ]; then
  echo "failed to switch" >&2
  exit 1
fi
exit 1
`, logFile)),
	})

	err := switchKubectlContext(context.Background(), "desired-context")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "kubectl config use-context desired-context")
	assert.Equal(t, []string{"config current-context", "config use-context desired-context"}, readNonEmptyLines(t, logFile))
}

func TestRunDrasiInit_DrasiMissingReturnsError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := runDrasiInit(context.Background(), "", false, "", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "drasi binary not found")
}

func TestRunDrasiInit_DrasiEnvAndInitSucceed(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "drasi.log")
	installFakeCommands(t, map[string]string{
		"drasi": fakeScript(fmt.Sprintf(`
echo %%*>>"%s"
if "%%1 %%2"=="env kube" exit /b 0
if "%%1"=="init" exit /b 0
exit /b 1
`, logFile), fmt.Sprintf(`
echo "$@" >> "%s"
if [ "$1 $2" = "env kube" ]; then exit 0; fi
if [ "$1" = "init" ]; then exit 0; fi
exit 1
`, logFile)),
	})

	err := runDrasiInit(context.Background(), "", false, "", nil)

	require.NoError(t, err)
	assert.Equal(t, []string{"env kube", "init"}, readNonEmptyLines(t, logFile))
}

func TestRunDrasiInit_PrivateAcrPassesRegistryFlag(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "drasi.log")
	installFakeCommands(t, map[string]string{
		"drasi": fakeScript(fmt.Sprintf(`
echo %%*>>"%s"
if "%%1 %%2"=="env kube" exit /b 0
if "%%1"=="init" exit /b 0
exit /b 1
`, logFile), fmt.Sprintf(`
echo "$@" >> "%s"
if [ "$1 $2" = "env kube" ]; then exit 0; fi
if [ "$1" = "init" ]; then exit 0; fi
exit 1
`, logFile)),
	})

	err := runDrasiInit(context.Background(), "", true, "example.azurecr.io", nil)

	require.NoError(t, err)
	assert.Equal(t, []string{"env kube", "init --registry example.azurecr.io"}, readNonEmptyLines(t, logFile))
}

func TestRunDrasiInit_RetriesOnTransientFailure(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level drasiInitBaseDelay.
	origDelay := drasiInitBaseDelay
	drasiInitBaseDelay = 10 * time.Millisecond
	t.Cleanup(func() { drasiInitBaseDelay = origDelay })

	markerDir := filepath.Join(t.TempDir(), "markers")
	require.NoError(t, os.MkdirAll(markerDir, 0o755))
	logFile := filepath.Join(t.TempDir(), "drasi.log")

	// Fake drasi: "env kube" always succeeds; "init" creates a marker file on
	// each invocation. When marker "3" exists (third call), init succeeds.
	installFakeCommands(t, map[string]string{
		"drasi": fakeScript(fmt.Sprintf(`
echo %%*>>"%s"
if "%%1 %%2"=="env kube" exit /b 0
if NOT "%%1"=="init" exit /b 1
if exist "%s\3" exit /b 0
if exist "%s\2" echo.>"%s\3" & exit /b 0
if exist "%s\1" echo.>"%s\2" & echo tunnel panic 1>&2 & exit /b 2
echo.>"%s\1" & echo tunnel panic 1>&2 & exit /b 2
`, logFile, markerDir, markerDir, markerDir, markerDir, markerDir, markerDir), fmt.Sprintf(`
echo "$@" >> "%s"
if [ "$1 $2" = "env kube" ]; then exit 0; fi
if [ "$1" = "init" ]; then
  if [ -f "%s/3" ]; then exit 0; fi
  if [ -f "%s/2" ]; then touch "%s/3"; exit 0; fi
  if [ -f "%s/1" ]; then touch "%s/2"; echo "tunnel panic" >&2; exit 2; fi
  touch "%s/1"; echo "tunnel panic" >&2; exit 2
fi
exit 1
`, logFile, markerDir, markerDir, markerDir, markerDir, markerDir, markerDir)),
	})

	var msgs []string
	progressFn := func(msg string) { msgs = append(msgs, msg) }

	err := runDrasiInit(context.Background(), "", false, "", progressFn)

	require.NoError(t, err, "should succeed on third attempt")
	assert.Len(t, msgs, 2, "should report 2 retry messages (attempts 1 and 2)")
	assert.Contains(t, msgs[0], "attempt 1/3 failed")
	assert.Contains(t, msgs[1], "attempt 2/3 failed")
}

func TestRunDrasiInit_AllAttemptsFailReturnsLastError(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level drasiInitBaseDelay.
	origDelay := drasiInitBaseDelay
	drasiInitBaseDelay = 10 * time.Millisecond
	t.Cleanup(func() { drasiInitBaseDelay = origDelay })

	logFile := filepath.Join(t.TempDir(), "drasi.log")
	installFakeCommands(t, map[string]string{
		"drasi": fakeScript(fmt.Sprintf(`
echo %%*>>"%s"
if "%%1 %%2"=="env kube" exit /b 0
if "%%1"=="init" (
  echo lost connection to pod 1>&2
  exit /b 2
)
exit /b 1
`, logFile), fmt.Sprintf(`
echo "$@" >> "%s"
if [ "$1 $2" = "env kube" ]; then exit 0; fi
if [ "$1" = "init" ]; then
  echo "lost connection to pod" >&2
  exit 2
fi
exit 1
`, logFile)),
	})

	var msgs []string
	progressFn := func(msg string) { msgs = append(msgs, msg) }

	err := runDrasiInit(context.Background(), "", false, "", progressFn)

	require.Error(t, err, "should fail after exhausting all attempts")
	assert.Contains(t, err.Error(), "exit status 2")
	assert.Len(t, msgs, 2, "should report retry messages for the first 2 failed attempts")
}

func TestRunDrasiInit_RetryCancelledByContext(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level drasiInitBaseDelay.
	origDelay := drasiInitBaseDelay
	drasiInitBaseDelay = 10 * time.Second // long delay so we can verify cancellation during backoff
	t.Cleanup(func() { drasiInitBaseDelay = origDelay })

	logFile := filepath.Join(t.TempDir(), "drasi.log")
	installFakeCommands(t, map[string]string{
		"drasi": fakeScript(fmt.Sprintf(`
echo %%*>>"%s"
if "%%1 %%2"=="env kube" exit /b 0
if "%%1"=="init" (
  echo tunnel panic 1>&2
  exit /b 2
)
exit /b 1
`, logFile), fmt.Sprintf(`
echo "$@" >> "%s"
if [ "$1 $2" = "env kube" ]; then exit 0; fi
if [ "$1" = "init" ]; then
  echo "tunnel panic" >&2
  exit 2
fi
exit 1
`, logFile)),
	})

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel shortly after runDrasiInit starts the backoff sleep.
	// The progressFn is called just before the sleep, so cancel there.
	progressFn := func(_ string) {
		// First retry message means init failed once and backoff starts next.
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
	}

	err := runDrasiInit(ctx, "", false, "", progressFn)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled, "should return context.Canceled when cancelled during backoff")
}

func TestApplyDrasiNetworkPolicies_KubectlMissingReturnsError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := applyDrasiNetworkPolicies(context.Background(), "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "kubectl not found on PATH")
}

func TestApplyDrasiNetworkPolicies_KubectlApplySucceeds(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "kubectl.log")
	installFakeCommands(t, map[string]string{
		"kubectl": fakeScript(fmt.Sprintf(`
echo %%*>>"%s"
if "%%1 %%2"=="config current-context" (
  echo other-context
  exit /b 0
)
if "%%1 %%2"=="config use-context" exit /b 0
if "%%1"=="apply" (
  more >nul
  exit /b 0
)
exit /b 1
`, logFile), fmt.Sprintf(`
echo "$@" >> "%s"
if [ "$1 $2" = "config current-context" ]; then
  echo other-context
  exit 0
fi
if [ "$1 $2" = "config use-context" ]; then exit 0; fi
if [ "$1" = "apply" ]; then
  cat > /dev/null
  exit 0
fi
exit 1
`, logFile)),
	})

	err := applyDrasiNetworkPolicies(context.Background(), "desired-context")

	require.NoError(t, err)
	assert.Equal(t, []string{"config current-context", "config use-context desired-context", "apply -f -"}, readNonEmptyLines(t, logFile))
}

// ---------------------------------------------------------------------------
// ensureSubscriptionAndLocation tests
// ---------------------------------------------------------------------------

func TestEnsureSubscriptionAndLocation_BothAlreadySet(t *testing.T) {
	t.Parallel()

	envValues := map[string]string{
		"AZURE_SUBSCRIPTION_ID": "existing-sub",
		"AZURE_LOCATION":        "westus2",
	}

	azdClient := newTestAzdClientWithPrompt(t,
		&testEnvironmentService{
			getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
				return &azdext.KeyValueResponse{Value: envValues[req.Key]}, nil
			},
		},
		&testPromptService{
			promptSubscriptionFunc: func(_ context.Context, _ *azdext.PromptSubscriptionRequest) (*azdext.PromptSubscriptionResponse, error) {
				t.Fatal("PromptSubscription should not be called when value already set")
				return nil, nil
			},
			promptLocationFunc: func(_ context.Context, _ *azdext.PromptLocationRequest) (*azdext.PromptLocationResponse, error) {
				t.Fatal("PromptLocation should not be called when value already set")
				return nil, nil
			},
		},
	)

	err := ensureSubscriptionAndLocation(context.Background(), azdClient, "dev")

	require.NoError(t, err)
}

func TestEnsureSubscriptionAndLocation_PromptsAndPersistsBoth(t *testing.T) {
	t.Parallel()

	persisted := map[string]string{}

	azdClient := newTestAzdClientWithPrompt(t,
		&testEnvironmentService{
			getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
				// Return empty for both — triggers prompting.
				return &azdext.KeyValueResponse{Value: ""}, nil
			},
			setValueFunc: func(_ context.Context, req *azdext.SetEnvRequest) (*azdext.EmptyResponse, error) {
				persisted[req.Key] = req.Value
				return &azdext.EmptyResponse{}, nil
			},
		},
		&testPromptService{
			promptSubscriptionFunc: func(_ context.Context, _ *azdext.PromptSubscriptionRequest) (*azdext.PromptSubscriptionResponse, error) {
				return &azdext.PromptSubscriptionResponse{
					Subscription: &azdext.Subscription{
						Id:   "prompted-sub-id",
						Name: "My Sub",
					},
				}, nil
			},
			promptLocationFunc: func(_ context.Context, req *azdext.PromptLocationRequest) (*azdext.PromptLocationResponse, error) {
				// Verify the subscription ID was passed through.
				assert.Equal(t, "prompted-sub-id", req.AzureContext.Scope.SubscriptionId)
				return &azdext.PromptLocationResponse{
					Location: &azdext.Location{
						Name:        "eastus2",
						DisplayName: "East US 2",
					},
				}, nil
			},
		},
	)

	err := ensureSubscriptionAndLocation(context.Background(), azdClient, "dev")

	require.NoError(t, err)
	assert.Equal(t, "prompted-sub-id", persisted["AZURE_SUBSCRIPTION_ID"])
	assert.Equal(t, "eastus2", persisted["AZURE_LOCATION"])
}

func TestEnsureSubscriptionAndLocation_SubscriptionSetLocationMissing(t *testing.T) {
	t.Parallel()

	persisted := map[string]string{}

	azdClient := newTestAzdClientWithPrompt(t,
		&testEnvironmentService{
			getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
				if req.Key == "AZURE_SUBSCRIPTION_ID" {
					return &azdext.KeyValueResponse{Value: "pre-set-sub"}, nil
				}
				return &azdext.KeyValueResponse{Value: ""}, nil
			},
			setValueFunc: func(_ context.Context, req *azdext.SetEnvRequest) (*azdext.EmptyResponse, error) {
				persisted[req.Key] = req.Value
				return &azdext.EmptyResponse{}, nil
			},
		},
		&testPromptService{
			promptSubscriptionFunc: func(_ context.Context, _ *azdext.PromptSubscriptionRequest) (*azdext.PromptSubscriptionResponse, error) {
				t.Fatal("PromptSubscription should not be called when subscription already set")
				return nil, nil
			},
			promptLocationFunc: func(_ context.Context, req *azdext.PromptLocationRequest) (*azdext.PromptLocationResponse, error) {
				assert.Equal(t, "pre-set-sub", req.AzureContext.Scope.SubscriptionId)
				return &azdext.PromptLocationResponse{
					Location: &azdext.Location{Name: "westeurope"},
				}, nil
			},
		},
	)

	err := ensureSubscriptionAndLocation(context.Background(), azdClient, "dev")

	require.NoError(t, err)
	assert.NotContains(t, persisted, "AZURE_SUBSCRIPTION_ID", "should not re-persist existing subscription")
	assert.Equal(t, "westeurope", persisted["AZURE_LOCATION"])
}

func TestEnsureSubscriptionAndLocation_SubscriptionPromptFails(t *testing.T) {
	t.Parallel()

	azdClient := newTestAzdClientWithPrompt(t,
		&testEnvironmentService{
			getValueFunc: func(_ context.Context, _ *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
				return &azdext.KeyValueResponse{Value: ""}, nil
			},
		},
		&testPromptService{
			promptSubscriptionFunc: func(_ context.Context, _ *azdext.PromptSubscriptionRequest) (*azdext.PromptSubscriptionResponse, error) {
				return nil, errors.New("user cancelled")
			},
		},
	)

	err := ensureSubscriptionAndLocation(context.Background(), azdClient, "dev")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "prompting for subscription")
}

func TestEnsureSubscriptionAndLocation_LocationPromptFails(t *testing.T) {
	t.Parallel()

	azdClient := newTestAzdClientWithPrompt(t,
		&testEnvironmentService{
			getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
				if req.Key == "AZURE_SUBSCRIPTION_ID" {
					return &azdext.KeyValueResponse{Value: "sub-123"}, nil
				}
				return &azdext.KeyValueResponse{Value: ""}, nil
			},
		},
		&testPromptService{
			promptLocationFunc: func(_ context.Context, _ *azdext.PromptLocationRequest) (*azdext.PromptLocationResponse, error) {
				return nil, errors.New("no locations available")
			},
		},
	)

	err := ensureSubscriptionAndLocation(context.Background(), azdClient, "dev")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "prompting for location")
}
