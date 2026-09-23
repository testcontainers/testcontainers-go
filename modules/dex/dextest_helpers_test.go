package dex_test

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// drivePasswordAuthCode exercises Dex's /dex/auth/local password connector.
// It performs: GET /dex/auth?... → follow to /dex/auth/local?req=...
// → parse the form action URL from the page body
// → POST credentials to the form action → follow the redirect to cfg.RedirectURL with ?code=... →
// exchange the code for a token.
//
// Returns the token response. Uses require (fatal) on protocol errors so
// callers don't need their own defensive checks.
func drivePasswordAuthCode(t *testing.T, ctx context.Context, cfg oauth2.Config, email, password string) *oauth2.Token {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			// Stop at the redirect back to our registered redirect URI —
			// we parse the code from that URL.
			if strings.HasPrefix(req.URL.String(), cfg.RedirectURL) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	authURL := cfg.AuthCodeURL("state-xyz", oauth2.AccessTypeOffline)

	// Step 1: GET /auth → follow redirects until the login form page.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	pageBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)

	// The form action is a relative path like /dex/auth/local/login?back=&state=<state>.
	// Extract it from the HTML rather than using resp.Request.URL — the page URL
	// (/dex/auth/local?req=...) and the form action (/dex/auth/local/login?...) differ.
	formAction := extractFormAction(t, string(pageBody))

	// Resolve the action against the base URL of the login page response.
	base := resp.Request.URL
	actionURL, err := base.Parse(formAction)
	require.NoError(t, err, "could not resolve form action %q against %q", formAction, base)

	// Step 2: POST the login form.
	form := url.Values{"login": {email}, "password": {password}}
	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, actionURL.String(), strings.NewReader(form.Encode()))
	require.NoError(t, err)
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	postResp, err := client.Do(postReq)
	require.NoError(t, err)
	defer postResp.Body.Close()
	_, _ = io.Copy(io.Discard, postResp.Body)

	// The code lands in the redirect URL. Depending on whether the client
	// stopped at a 302 Location or followed through to the redirect URI as
	// the final request, check both.
	var codeLoc *url.URL
	if postResp.Request.URL != nil && strings.HasPrefix(postResp.Request.URL.String(), cfg.RedirectURL) {
		codeLoc = postResp.Request.URL
	} else if loc := postResp.Header.Get("Location"); loc != "" {
		parsed, perr := url.Parse(loc)
		require.NoError(t, perr)
		codeLoc = parsed
	}
	require.NotNil(t, codeLoc, "expected redirect to %q with ?code=…, got final URL %q",
		cfg.RedirectURL, postResp.Request.URL)

	code := codeLoc.Query().Get("code")
	require.NotEmpty(t, code, "no ?code= in redirect URL %q", codeLoc.String())

	// Step 3: Exchange the code for a token using a context-aware HTTP client
	// so the exchange honours the test deadline.
	httpCtxClient := &http.Client{}
	tokenCtx := context.WithValue(ctx, oauth2.HTTPClient, httpCtxClient)
	tok, err := cfg.Exchange(tokenCtx, code)
	require.NoError(t, err, "token exchange")
	return tok
}

// extractFormAction parses the value of the first <form ... action="..."> attribute
// from an HTML page body. Uses simple string search — sufficient for Dex's
// single-form login page without pulling in golang.org/x/net/html.
func extractFormAction(t *testing.T, body string) string {
	t.Helper()
	_, rest, ok := strings.Cut(body, `action="`)
	require.True(t, ok, "could not find form action in login page HTML")
	raw, _, ok := strings.Cut(rest, `"`)
	require.True(t, ok, "malformed form action attribute in login page HTML")
	// HTML-unescape &amp; → &
	return strings.ReplaceAll(raw, "&amp;", "&")
}
