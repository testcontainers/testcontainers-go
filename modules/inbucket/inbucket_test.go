package inbucket

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestInbucket(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "inbucket/inbucket:sha-2d409bb")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// smtpConnection {
	smtpURL, err := ctr.SmtpConnection(ctx)
	// }
	require.NoError(t, err)

	// webInterface {
	webInterfaceURL, err := ctr.WebInterface(ctx)
	// }
	require.NoError(t, err)
	httpClient := &http.Client{Timeout: 30 * time.Second}
	listMailbox := func() []struct {
		ID string `json:"id"`
	} {
		t.Helper()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet,
			webInterfaceURL+"/api/v1/mailbox/"+url.QueryEscape("to@example.org"), nil)
		require.NoError(t, err)

		response, err := httpClient.Do(request)
		require.NoError(t, err)
		defer response.Body.Close()

		const maxBody = 1 << 20
		body, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
		require.NoError(t, err)
		require.LessOrEqual(t, len(body), maxBody, "mailbox response exceeds limit")
		require.Equal(t, http.StatusOK, response.StatusCode, "%s", body)

		var headers []struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(body, &headers))
		return headers
	}

	headers := listMailbox()
	require.Empty(t, headers)

	msg := []byte("To: to@example.org\r\n" +
		"Subject: Testcontainers test!\r\n" +
		"\r\n" +
		"This is a Testcontainers test.\r\n")
	err = smtp.SendMail(smtpURL, nil, "from@example.org", []string{"to@example.org"}, msg)
	require.NoError(t, err)

	headers = listMailbox()
	require.Len(t, headers, 1)
}
