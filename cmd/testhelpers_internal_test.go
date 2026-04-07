package cmd

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type testEnvironmentService struct {
	azdext.UnimplementedEnvironmentServiceServer
	getCurrentFunc func(context.Context, *azdext.EmptyRequest) (*azdext.EnvironmentResponse, error)
	getValueFunc   func(context.Context, *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error)
	setValueFunc   func(context.Context, *azdext.SetEnvRequest) (*azdext.EmptyResponse, error)
}

func (s *testEnvironmentService) GetCurrent(
	ctx context.Context,
	req *azdext.EmptyRequest,
) (*azdext.EnvironmentResponse, error) {
	if s.getCurrentFunc != nil {
		return s.getCurrentFunc(ctx, req)
	}
	return nil, nil
}

func (s *testEnvironmentService) GetValue(
	ctx context.Context,
	req *azdext.GetEnvRequest,
) (*azdext.KeyValueResponse, error) {
	if s.getValueFunc != nil {
		return s.getValueFunc(ctx, req)
	}
	return nil, nil
}

func (s *testEnvironmentService) SetValue(
	ctx context.Context,
	req *azdext.SetEnvRequest,
) (*azdext.EmptyResponse, error) {
	if s.setValueFunc != nil {
		return s.setValueFunc(ctx, req)
	}
	return &azdext.EmptyResponse{}, nil
}

func startTestEnvironmentServer(t *testing.T, service *testEnvironmentService) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	azdext.RegisterEnvironmentServiceServer(server, service)

	go func() {
		_ = server.Serve(listener)
	}()

	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})

	return listener.Addr().String()
}

func newTestAzdClient(t *testing.T, service *testEnvironmentService) *azdext.AzdClient {
	t.Helper()

	client, err := azdext.NewAzdClient(azdext.WithAddress(startTestEnvironmentServer(t, service)))
	require.NoError(t, err)
	t.Cleanup(client.Close)

	return client
}

func installFakeCommands(t *testing.T, scripts map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, script := range scripts {
		path := filepath.Join(dir, name+".cmd")
		content := "@echo off\r\n" + script + "\r\n"
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return dir
}

func readNonEmptyLines(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		filtered = append(filtered, trimmed)
	}

	return filtered
}
