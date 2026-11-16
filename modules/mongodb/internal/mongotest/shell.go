package mongotest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

const resultMarker = "TC_MONGO_RESULT "

// Shell owns an official mongo:7 tool container and its test-process relay.
// No target-image shell, host-installed tool, local-service shortcut or write retry is used.
type Shell struct {
	ctx     context.Context
	helper  testcontainers.Container
	uri     string
	secrets []string
}

func New(t *testing.T, parent context.Context, rawURI string) *Shell {
	t.Helper()
	ctx, cancel := context.WithTimeout(parent, 5*time.Minute)
	t.Cleanup(cancel)
	u, err := url.Parse(rawURI)
	require.NoError(t, err, "parse generated MongoDB URI")
	r, err := newRelay(ctx, u.Host)
	require.NoError(t, err, "start MongoDB endpoint relay")
	t.Cleanup(r.close)
	port := r.listener.Addr().(*net.TCPAddr).Port
	uri, err := relayURI(rawURI, port)
	require.NoError(t, err)
	helper, err := testcontainers.Run(ctx, "mongo:7",
		testcontainers.WithEntrypoint("/bin/sh"),
		testcontainers.WithCmd("-c", "sleep 300"),
		testcontainers.WithHostPortAccess(port))
	testcontainers.CleanupContainer(t, helper)
	require.NoError(t, err, "start official mongosh helper")
	secrets := []string{rawURI, uri}
	if u.User != nil {
		password, _ := u.User.Password()
		if password != "" {
			secrets = append(secrets, password, url.QueryEscape(password))
		}
	}
	return &Shell{ctx: ctx, helper: helper, uri: uri, secrets: secrets}
}

// relayURI only handles the module's generated single-host TCP URIs. Keep the
// original credentials, database, and query; direct mode avoids topology discovery
// escaping the fixed mapped-endpoint transport.
func relayURI(raw string, port int) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("invalid generated MongoDB URI")
	}
	if u.Scheme != "mongodb" {
		return "", errors.New("expected generated mongodb URI")
	}
	host, originalPort, err := net.SplitHostPort(u.Host)
	if err != nil || host == "" || originalPort == "" || strings.Contains(host, ",") {
		return "", errors.New("expected generated single host and mapped port")
	}
	u.Host = net.JoinHostPort(testcontainers.HostInternal, strconv.Itoa(port))
	q := u.Query()
	q.Set("directConnection", "true")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// Run binds the URI as a JSON string in a copied JS file, never shell text or
// credential-bearing argv. Docker execution, the shell process, and the relay all
// have deadlines. A failed operation is reported once, without a fallback client.
// The body must return a JSON-compatible result; BSON documents should first be
// serialized with EJSON.stringify(value, null, 0, {relaxed:false}).
func (s *Shell) Run(parent context.Context, body string, result any) error {
	ctx, cancel := context.WithTimeout(parent, 2*time.Minute)
	defer cancel()
	stop := context.AfterFunc(s.ctx, cancel)
	defer stop()
	program := shellProgram(s.uri, body)
	if err := s.helper.CopyToContainer(ctx, program, "/tmp/tc-mongo-operation.js", 0o600); err != nil {
		return s.safeError("copy mongosh operation", err)
	}
	code, reader, err := s.helper.Exec(ctx, []string{"timeout", "--signal=KILL", "125s", "mongosh", "--quiet", "--norc", "--nodb", "--file", "/tmp/tc-mongo-operation.js"}, tcexec.Multiplexed())
	if err != nil {
		return s.safeError("execute mongosh", err)
	}
	if reader == nil {
		return errors.New("mongosh returned no output reader")
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		return s.safeError("read mongosh output", err)
	}
	if code != 0 {
		return s.safeError(fmt.Sprintf("mongosh exit %d", code), errors.New(string(output)))
	}
	return decodeResult(output, result)
}

func shellProgram(uri, body string) []byte {
	bound, _ := json.Marshal(uri)
	return []byte(`const connection = new Mongo(` + string(bound) + `);
const db = connection.getDB('test');
function commandWorked(result) { if (result.ok !== 1) throw new Error(EJSON.stringify(result)); return result; }
function acknowledged(result) { if (result.acknowledged !== true) throw new Error('write was not acknowledged'); return result; }
try {
 const result = (() => {
` + body + `
 })();
 print('` + resultMarker + `' + JSON.stringify(result));
} finally { connection.close(); }
`)
}

func (s *Shell) safeError(operation string, err error) error {
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%s: %w", operation, context.Canceled)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: %w", operation, context.DeadlineExceeded)
	}
	message := err.Error()
	for _, secret := range s.secrets {
		message = strings.ReplaceAll(message, secret, "[redacted]")
	}
	return fmt.Errorf("%s: %s", operation, message)
}

func decodeResult(output []byte, result any) error {
	var payload []byte
	for _, line := range bytes.Split(output, []byte("\n")) {
		if bytes.HasPrefix(line, []byte(resultMarker)) {
			if payload != nil {
				return errors.New("multiple mongosh result markers")
			}
			payload = bytes.TrimPrefix(line, []byte(resultMarker))
		}
	}
	if payload == nil {
		return errors.New("missing mongosh result marker")
	}
	if result == nil {
		if !json.Valid(payload) {
			return errors.New("invalid mongosh result JSON")
		}
		return nil
	}
	if err := json.Unmarshal(payload, result); err != nil {
		return fmt.Errorf("decode mongosh result: %w", err)
	}
	return nil
}

// CanonicalDocument compacts canonical EJSON without decoding numbers or changing
// field order. Equality therefore preserves both BSON numeric types and bson.D order.
func CanonicalDocument(raw string) (string, error) {
	var compact bytes.Buffer
	if err := json.Compact(&compact, []byte(raw)); err != nil {
		return "", err
	}
	return compact.String(), nil
}
