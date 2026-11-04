package mongotest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

func TestRelayURIPreservesFields(t *testing.T) {
	for _, authority := range []string{"original.example:32123", "[2001:db8::1]:32123"} {
		t.Run(authority, func(t *testing.T) {
			original := &url.URL{Scheme: "mongodb", Host: authority, User: url.UserPassword("user@name", "p@ss:'\"\\$()`\n"), Path: "/original-db", RawQuery: "authSource=admin&replicaSet=rs&appName=a%2Bb&directConnection=true"}
			raw, err := relayURI(original.String(), 54321)
			require.NoError(t, err)
			got, err := url.Parse(raw)
			require.NoError(t, err)
			require.Equal(t, "host.testcontainers.internal:54321", got.Host)
			require.Equal(t, original.User, got.User)
			require.Equal(t, original.Path, got.Path)
			require.Equal(t, original.Query(), got.Query())
			require.Equal(t, authority, original.Host)
			// A credential containing executable-looking text is one JS string literal.
			program := string(shellProgram(raw, "return true;"))
			literal := strings.SplitN(strings.TrimPrefix(program, "const connection = new Mongo("), ");\n", 2)[0]
			var bound string
			require.NoError(t, json.Unmarshal([]byte(literal), &bound))
			require.Equal(t, raw, bound)
		})
	}
	for _, raw := range []string{"mongodb+srv://example/", "mongodb://host/", "mongodb://a:1,b:2/", "mongodb://:123/", "mongodb://host:bad/"} {
		_, err := relayURI(raw, 54321)
		require.Error(t, err)
	}
	raw, err := relayURI("mongodb://original.example:32123/?replicaSet=rs", 54321)
	require.NoError(t, err)
	got, err := url.Parse(raw)
	require.NoError(t, err)
	require.Equal(t, "true", got.Query().Get("directConnection"))
	require.Equal(t, "rs", got.Query().Get("replicaSet"))
}

// Only the methods used by Run are replaced. No Docker client is initialized.
type shellContainer struct {
	testcontainers.Container
	calls   int
	code    int
	output  string
	execErr error
	copyErr error
	argv    []string
	program []byte
}

func (c *shellContainer) CopyToContainer(_ context.Context, program []byte, _ string, _ int64) error {
	c.program = append([]byte(nil), program...)
	return c.copyErr
}

func (c *shellContainer) Exec(ctx context.Context, argv []string, _ ...tcexec.ProcessOption) (int, io.Reader, error) {
	c.calls++
	c.argv = append([]string(nil), argv...)
	if _, ok := ctx.Deadline(); !ok {
		return 0, nil, errors.New("missing operation deadline")
	}
	return c.code, strings.NewReader(c.output), c.execErr
}

func TestShellErrorsAndBinding(t *testing.T) {
	for _, tc := range []struct {
		name             string
		code             int
		output           string
		execErr, copyErr error
		want             string
		calls            int
	}{
		{name: "success", output: resultMarker + `{"acknowledged":true}` + "\n", calls: 1},
		{name: "official shell failure after result", code: 1, output: resultMarker + `{"acknowledged":true}` + "\nMongoServerError: Authentication failed. password-secret", want: "mongosh exit 1", calls: 1},
		{name: "docker exec failure", execErr: errors.New("exec password-secret"), want: "execute mongosh", calls: 1},
		{name: "copy failure", copyErr: errors.New("copy password-secret"), want: "copy mongosh operation"},
		{name: "missing result", output: "MongoServerError", want: "missing mongosh result marker", calls: 1},
		{name: "duplicate result", output: resultMarker + "{}\n" + resultMarker + "{}\n", want: "multiple mongosh result markers", calls: 1},
		{name: "invalid result", output: resultMarker + "{\n", want: "decode mongosh result", calls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &shellContainer{code: tc.code, output: tc.output, execErr: tc.execErr, copyErr: tc.copyErr}
			s := &Shell{ctx: context.Background(), helper: fake, uri: "mongodb://user:password-secret@host:123/db", secrets: []string{"password-secret"}}
			var result struct{ Acknowledged bool }
			err := s.Run(context.Background(), "return {acknowledged:true};", &result)
			if tc.want == "" {
				require.NoError(t, err)
				require.True(t, result.Acknowledged)
			} else {
				require.ErrorContains(t, err, tc.want)
				require.NotContains(t, err.Error(), "password-secret")
			}
			require.Equal(t, tc.calls, fake.calls, "failed writes must never be retried")
			require.NotContains(t, strings.Join(fake.argv, " "), "password-secret")
			require.Contains(t, string(fake.program), "password-secret")
		})
	}
	// Startup chatter is ignored, but a result must decode deterministically.
	var result int
	require.NoError(t, decodeResult([]byte("startup\n"+resultMarker+"1\n"), &result))
	require.Equal(t, 1, result)
}

func TestCanonicalDocumentTypesAndOrder(t *testing.T) {
	original, err := CanonicalDocument(`{ "_id": { "$numberInt": "1" }, "seeded": true }`)
	require.NoError(t, err)
	require.Equal(t, `{"_id":{"$numberInt":"1"},"seeded":true}`, original) //nolint:testifylint // BSON document field order is part of the assertion.
	for _, other := range []string{`{"_id":{"$numberLong":"1"},"seeded":true}`, `{"_id":{"$numberDouble":"1.0"},"seeded":true}`, `{"seeded":true,"_id":{"$numberInt":"1"}}`, `{"_id":{"$numberLong":"9007199254740993"},"seeded":true}`} {
		got, err := CanonicalDocument(other)
		require.NoError(t, err)
		require.NotEqual(t, original, got)
		require.Equal(t, other, got)
	}
	_, err = CanonicalDocument("{")
	require.Error(t, err)
}

func TestShellOperationCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fake := &cancelContainer{}
	s := &Shell{ctx: ctx, helper: fake, uri: "mongodb://host:123"}
	started := time.Now()
	require.ErrorIs(t, s.Run(context.Background(), "return true;", nil), context.Canceled)
	require.Less(t, time.Since(started), time.Second)
}

type cancelContainer struct{ testcontainers.Container }

func (*cancelContainer) CopyToContainer(ctx context.Context, _ []byte, _ string, _ int64) error {
	<-ctx.Done()
	return ctx.Err()
}
