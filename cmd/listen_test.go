package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/azure/azd.extensions.drasi/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func saveListenVars(t *testing.T) {
	t.Helper()
	origDrasiCheckFunc := drasiCheckFunc
	origPreDeployValidateFunc := preDeployValidateFunc
	origPostProvisionTimeout := postProvisionTimeout
	origPostProvisionInterval := postProvisionInterval
	t.Cleanup(func() {
		drasiCheckFunc = origDrasiCheckFunc
		preDeployValidateFunc = origPreDeployValidateFunc
		postProvisionTimeout = origPostProvisionTimeout
		postProvisionInterval = origPostProvisionInterval
	})
}

func TestHandlePostProvision_Ready(t *testing.T) {
	saveListenVars(t)

	drasiCheckFunc = func(context.Context) error { return nil }

	err := handlePostProvision(context.Background(), nil)
	require.NoError(t, err)
}

func TestWaitForDrasiReady_Timeout_ReturnsError(t *testing.T) {
	saveListenVars(t)

	drasiCheckFunc = func(context.Context) error { return errors.New("not ready") }

	err := waitForDrasiReady(context.Background(), 50*time.Millisecond, 10*time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "drasi API not ready")
}

func TestHandlePostProvision_Timeout_NonBlocking(t *testing.T) {
	saveListenVars(t)

	drasiCheckFunc = func(context.Context) error { return errors.New("not ready") }
	postProvisionTimeout = 50 * time.Millisecond
	postProvisionInterval = 10 * time.Millisecond

	err := handlePostProvision(context.Background(), nil)
	require.NoError(t, err)
}

func TestHandlePreDeploy_Valid(t *testing.T) {
	saveListenVars(t)

	preDeployValidateFunc = func(dir, file, env string) (*validation.ValidationResult, error) {
		assert.Equal(t, "drasi", dir)
		assert.Equal(t, "drasi.yaml", file)
		assert.Empty(t, env)
		return &validation.ValidationResult{}, nil
	}

	err := handlePreDeploy(context.Background(), nil)
	require.NoError(t, err)
}

func TestHandlePreDeploy_ValidationErrors_BlocksDeploy(t *testing.T) {
	saveListenVars(t)

	preDeployValidateFunc = func(dir, file, env string) (*validation.ValidationResult, error) {
		return &validation.ValidationResult{
			Issues: []validation.ValidationIssue{{
				Level:   validation.LevelError,
				Message: "missing required field",
			}},
		}, nil
	}

	err := handlePreDeploy(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pre-deploy validation failed")
}

func TestHandlePreDeploy_WarningsOnly_Proceeds(t *testing.T) {
	saveListenVars(t)

	preDeployValidateFunc = func(dir, file, env string) (*validation.ValidationResult, error) {
		return &validation.ValidationResult{
			Issues: []validation.ValidationIssue{{
				Level:   validation.LevelWarning,
				Message: "optional field missing",
			}},
		}, nil
	}

	err := handlePreDeploy(context.Background(), nil)
	require.NoError(t, err)
}

func TestHandlePreDeploy_ManifestLoadError(t *testing.T) {
	saveListenVars(t)

	preDeployValidateFunc = func(dir, file, env string) (*validation.ValidationResult, error) {
		return nil, errors.New("file not found")
	}

	err := handlePreDeploy(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pre-deploy validation")
}
