package cmd_test

import (
	"bytes"
	"testing"

	"github.com/azure/azd.extensions.drasi/cmd"
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

// TestProvisionCommand_EnvironmentFlag_Accepted verifies that --environment is registered
// on the provision command.
func TestProvisionCommand_EnvironmentFlag_Accepted(t *testing.T) {
	_, provision := getProvisionCommand(t)
	assert.NotNil(t, provision.Flags().Lookup("environment"), "--environment must be a registered flag on provision")
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
