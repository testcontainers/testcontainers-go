package testcbt

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

const lookupOutput = "----------------------------------------\n\"1\"\n  \"name:firstName\"                         @ 2026/11/04-10:35:42.123000\n    \"Gopher\"\n\n"

func TestLookupCell(t *testing.T) {
	row, column, value, err := LookupCell(lookupOutput)
	require.NoError(t, err)
	require.Equal(t, "1", row)
	require.Equal(t, "name:firstName", column)
	require.Equal(t, "Gopher", value)

	_, _, value, err = LookupCell(strings.Replace(lookupOutput, `"Gopher"`, `"Gopher\n\"quoted\"\x00"`, 1))
	require.NoError(t, err)
	require.Equal(t, "Gopher\n\"quoted\"\x00", value)

	for name, output := range map[string]string{
		"missing row":     "",
		"missing cell":    "----------------------------------------\n\"1\"\n",
		"diagnostic only": "Reading row: Gopher",
		"bad row quoting": strings.Replace(lookupOutput, `"1"`, "1", 1),
		"bad value":       strings.Replace(lookupOutput, `"Gopher"`, `"Gopher\z"`, 1),
		"bad timestamp":   strings.Replace(lookupOutput, "2026/11/04", "not-a-date", 1),
		"extra record":    lookupOutput + "unexpected\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, _, _, err := LookupCell(output)
			require.Error(t, err)
		})
	}
}

type fakeExecutor struct {
	code   int
	reader io.Reader
	err    error
	args   []string
	env    []string
}

func (f *fakeExecutor) Exec(ctx context.Context, args []string, options ...tcexec.ProcessOption) (int, io.Reader, error) {
	if _, ok := ctx.Deadline(); !ok {
		return 0, nil, errors.New("missing deadline")
	}
	f.args = args
	opts := tcexec.NewProcessOptions(args)
	for _, option := range options {
		option.Apply(opts)
	}
	f.env = opts.ExecConfig.Env
	return f.code, f.reader, f.err
}

func TestRunErrorsAndArgumentBinding(t *testing.T) {
	const project = "project;$(never-run)"
	fake := &fakeExecutor{reader: strings.NewReader(lookupOutput)}
	output, err := Run(context.Background(), fake, "host:1234", project, "instance", "lookup", "table", "1", "columns=name:")
	require.NoError(t, err)
	require.Equal(t, lookupOutput, output)
	require.Equal(t, []string{"/bin/sh", "-c", boundedCommand, "cbt-test"}, fake.args[:4])
	require.Equal(t, []string{
		"-project=" + project, "-instance=instance", "-timeout=20s",
		"-admin-endpoint=host:1234", "-data-endpoint=host:1234",
		"-creds=/tmp/unused-emulator-credentials", "-cert-file=", "-access-token=", "-auth-token=",
		"lookup", "table", "1", "columns=name:",
	}, fake.args[4:])
	require.Equal(t, []string{"BIGTABLE_EMULATOR_HOST=host:1234", "TZ=UTC"}, fake.env)

	for name, failure := range map[string]*fakeExecutor{
		"transport": {err: errors.New("attach failed")},
		"exit":      {code: 1, reader: strings.NewReader("Reading row: denied")},
		"read":      {reader: errorReader{}},
		"oversized": {reader: strings.NewReader(strings.Repeat("x", (128<<10)+1))},
		"no reader": {},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Run(context.Background(), failure, "host:1234", project, "instance", "lookup", "table", "1", "columns=name:")
			require.Error(t, err)
			if name == "exit" {
				require.ErrorContains(t, err, "exit 1: Reading row: denied")
			}
		})
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("broken output") }
