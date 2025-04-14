package inbucket

import (
	"context"
	"net/smtp"
	"testing"

	"github.com/inbucket/inbucket/pkg/rest/client"
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
	restClient, err := client.New(webInterfaceURL)
	require.NoError(t, err)

	headers, err := restClient.ListMailbox("to@example.org")
	require.NoError(t, err)
	require.Empty(t, headers)

	msg := []byte("To: to@example.org\r\n" +
		"Subject: Testcontainers test!\r\n" +
		"\r\n" +
		"This is a Testcontainers test.\r\n")
	err = smtp.SendMail(smtpURL, nil, "from@example.org", []string{"to@example.org"}, msg)
	require.NoError(t, err)

	headers, err = restClient.ListMailbox("to@example.org")
	require.NoError(t, err)
	require.Len(t, headers, 1)
}
