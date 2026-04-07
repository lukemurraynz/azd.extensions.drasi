package drasi

import (
	"context"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListComponents_ParsesPipeDelimitedSourceOutput(t *testing.T) {
	t.Parallel()

	// Real drasi CLI output for `drasi list source`.
	stdout := "      ID     | AVAILABLE | INGRESS URL | MESSAGES  \n" +
		"-------------+-----------+-------------+-----------\n" +
		"  k8s-source | true      |             |\n"

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: stdout}}})

	got, err := client.ListComponents(context.Background(), "source")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "k8s-source", got[0].ID)
	assert.Equal(t, "source", got[0].Kind)
	assert.Equal(t, "true", got[0].Status)
}

func TestListComponents_ParsesPipeDelimitedQueryOutput(t *testing.T) {
	t.Parallel()

	// Real drasi CLI output for `drasi list query`.
	stdout := "        ID       | CONTAINER | ERRORMESSAGE |              HOSTNAME               | STATUS   \n" +
		"-----------------+-----------+--------------+-------------------------------------+----------\n" +
		"  k8s-pods-query | default   |              | default-query-host-847db9c696-grhvc | Running\n"

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: stdout}}})

	got, err := client.ListComponents(context.Background(), "query")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "k8s-pods-query", got[0].ID)
	assert.Equal(t, "query", got[0].Kind)
	assert.Equal(t, "Running", got[0].Status)
}

func TestListComponents_ParsesPipeDelimitedReactionOutput(t *testing.T) {
	t.Parallel()

	// Real drasi CLI output for `drasi list reaction`.
	stdout := "        ID       | AVAILABLE | INGRESS URL | MESSAGES  \n" +
		"-----------------+-----------+-------------+-----------\n" +
		"  debug-reaction | true      |             |\n"

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: stdout}}})

	got, err := client.ListComponents(context.Background(), "reaction")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "debug-reaction", got[0].ID)
	assert.Equal(t, "reaction", got[0].Kind)
	assert.Equal(t, "true", got[0].Status)
}

func TestListComponents_PipeDelimited_MultipleRows(t *testing.T) {
	t.Parallel()

	stdout := "  ID       | AVAILABLE | INGRESS URL | MESSAGES  \n" +
		"-----------+-----------+-------------+-----------\n" +
		"  source-a | true      |             |\n" +
		"  source-b | false     | http://x    | degraded\n"

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: stdout}}})

	got, err := client.ListComponents(context.Background(), "source")

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "source-a", got[0].ID)
	assert.Equal(t, "true", got[0].Status)
	assert.Equal(t, "source-b", got[1].ID)
	assert.Equal(t, "false", got[1].Status)
}

func TestListComponents_ParsesLegacyTabularOutput(t *testing.T) {
	t.Parallel()

	// Legacy space-delimited format with ID KIND STATUS columns.
	stdout := "ID KIND STATUS\nalerts query Online\norders query Pending\n"

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: stdout}}})

	got, err := client.ListComponents(context.Background(), "query")

	require.NoError(t, err)
	assert.Equal(t, []ComponentSummary{
		{ID: "alerts", Kind: "query", Status: "Online"},
		{ID: "orders", Kind: "query", Status: "Pending"},
	}, got)
}

func TestListComponents_LegacyHeaderWithExtraSpacing(t *testing.T) {
	t.Parallel()

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: "\nID   KIND    STATUS\nalerts query Online\n"}}})

	got, err := client.ListComponents(context.Background(), "query")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, ComponentSummary{ID: "alerts", Kind: "query", Status: "Online"}, got[0])
}

func TestListComponents_LegacyStatusWithSpaces_Joined(t *testing.T) {
	t.Parallel()

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: "ID KIND STATUS\nalerts query Not Ready\n"}}})

	got, err := client.ListComponents(context.Background(), "query")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Not Ready", got[0].Status)
}

func TestListComponentsInContext_PassesContextFlag(t *testing.T) {
	t.Parallel()

	stdout := "  ID     | AVAILABLE | INGRESS URL | MESSAGES  \n" +
		"---------+-----------+-------------+-----------\n" +
		"  src1   | true      |             |\n"
	runner := &mockRunner{responses: []runnerResponse{{stdout: stdout}}}
	client := newTestClient(runner)

	got, err := client.ListComponentsInContext(context.Background(), "source", "aks-dev")

	require.NoError(t, err)
	require.Len(t, got, 1)
	require.NotEmpty(t, runner.args)
	assert.Equal(t, []string{"--context", "aks-dev", "list", "source"}, runner.args[0])
}

func TestListComponentsInContext_UnknownContextFlag_FallsBackWithoutContext(t *testing.T) {
	t.Parallel()

	stdout := "  ID     | AVAILABLE | INGRESS URL | MESSAGES  \n" +
		"---------+-----------+-------------+-----------\n" +
		"  src1   | true      |             |\n"
	runner := &mockRunner{responses: []runnerResponse{
		{stderr: "Error: unknown flag: --context", exitCode: 1},
		{stdout: stdout},
	}}
	client := newTestClient(runner)

	got, err := client.ListComponentsInContext(context.Background(), "source", "aks-dev")

	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Len(t, runner.args, 2)
	assert.Equal(t, []string{"--context", "aks-dev", "list", "source"}, runner.args[0])
	assert.Equal(t, []string{"list", "source"}, runner.args[1])
}

func TestListComponents_EmptyList_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		stdout string
	}{
		{name: "pipe delimited with header only", stdout: "  ID | AVAILABLE | INGRESS URL\n------+-----------+------------\n"},
		{name: "legacy header only", stdout: "ID KIND STATUS\n"},
		{name: "empty output", stdout: ""},
		{name: "whitespace only", stdout: "  \n  \n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: tt.stdout}}})

			got, err := client.ListComponents(context.Background(), "source")

			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Empty(t, got)
		})
	}
}

func TestListComponents_NoResourcesMessage_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: "No query components found.\n"}}})

	got, err := client.ListComponents(context.Background(), "query")

	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestListComponents_NonZeroExit_MapsToCliError(t *testing.T) {
	t.Parallel()

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stderr: "connection failed", exitCode: 1}}})

	got, err := client.ListComponents(context.Background(), "query")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), output.ERR_DRASI_CLI_ERROR)
}

func TestListComponents_ErrorWrittenToStdout_MapsToCliError(t *testing.T) {
	t.Parallel()

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: "Error: network unreachable", exitCode: 0}}})

	got, err := client.ListComponents(context.Background(), "query")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), output.ERR_DRASI_CLI_ERROR)
}

func TestListComponents_LegacyMalformedHeader_ReturnsError(t *testing.T) {
	t.Parallel()

	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: "not a drasi table"}}})

	got, err := client.ListComponents(context.Background(), "query")

	require.Error(t, err)
	assert.Nil(t, got)
}

func TestListComponents_PipeDelimited_MissingIDColumn_ReturnsError(t *testing.T) {
	t.Parallel()

	stdout := "  NAME | AVAILABLE | STATUS\n" +
		"-------+-----------+-------\n" +
		"  foo  | true      | ok\n"
	client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: stdout}}})

	got, err := client.ListComponents(context.Background(), "source")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "missing ID column")
}

func TestParseListOutput_QueryStatusPreferredOverAvailable(t *testing.T) {
	t.Parallel()

	// If both AVAILABLE and STATUS columns exist, STATUS wins for queries.
	stdout := "  ID     | AVAILABLE | STATUS\n" +
		"---------+-----------+--------\n" +
		"  q1     | true      | Running\n"

	got, err := parseListOutput(stdout, "query")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Running", got[0].Status)
}

func TestParseListOutput_SourceAvailableUsedWhenNoStatusColumn(t *testing.T) {
	t.Parallel()

	stdout := "  ID     | AVAILABLE | INGRESS URL\n" +
		"---------+-----------+------------\n" +
		"  src1   | true      | \n"

	got, err := parseListOutput(stdout, "source")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "true", got[0].Status)
}
