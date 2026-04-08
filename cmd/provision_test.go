package cmd_test

import (
	"bytes"
	yaml "gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/cmd"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T041: provision command CLI flag-registration tests.
//
// These black-box tests confirm that flags are declared correctly on the provision
// command. They do not require a live Azure connection and pass regardless of the
// underlying provision implementation.
//
// Success-path and injection tests live in provision_internal_test.go (package cmd)
// because they override the package-level runProvisionFunc variable.

func getProvisionCommand(t *testing.T) (*cobra.Command, *cobra.Command) {
	t.Helper()
	root := cmd.NewRootCommand()
	for _, c := range root.Commands() {
		if c.Name() == "provision" {
			return root, c
		}
	}
	t.Fatalf("provision command not found on root")
	return nil, nil
}

// TestProvisionCommand_EnvironmentFlag_Accepted verifies that --environment is available
// to the provision command via the root persistent flag (not a local flag).
func TestProvisionCommand_EnvironmentFlag_Accepted(t *testing.T) {
	root, provision := getProvisionCommand(t)
	// --environment is a root persistent flag inherited by all subcommands, not a local flag on provision.
	flag := provision.InheritedFlags().Lookup("environment")
	if flag == nil {
		flag = root.PersistentFlags().Lookup("environment")
	}
	assert.NotNil(t, flag, "--environment must be available to provision via root persistent flags")
}

// TestProvisionCommand_OutputJSONFlag_Accepted verifies that --output json is accepted
// by the root command and inherited by provision.
func TestProvisionCommand_OutputJSONFlag_Accepted(t *testing.T) {
	root, _ := getProvisionCommand(t)
	assert.NotNil(t, root.PersistentFlags().Lookup("output"), "--output must be a root persistent flag accepted by provision")
}

// TestProvisionCommand_NoAuth_ExitsTwo verifies the provision command fails fast
// with ERR_NO_AUTH when AZD_SERVER is not configured.
func TestProvisionCommand_NoAuth_ExitsTwo(t *testing.T) {
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_NO_AUTH", "provision must return ERR_NO_AUTH without AZD_SERVER")
}

// TestNetworkPolicyYAMLValid ensures the embedded network policy YAML is parseable
// and contains at least one document.
func TestNetworkPolicyYAMLValid(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "cmd", "network_policies.yaml")
	// Fallback to module-relative path if test runs from repo root
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join("cmd", "network_policies.yaml")
	}
	data, err := os.ReadFile(path)
	require.NoError(t, err, "reading network_policies.yaml should succeed")
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var doc interface{}
	count := 0
	for {
		if err := dec.Decode(&doc); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("failed to parse YAML document %d: %v", count+1, err)
		}
		count++
	}
	require.Equal(t, 8, count, "expected exactly 8 NetworkPolicy YAML documents in network_policies.yaml")
}
