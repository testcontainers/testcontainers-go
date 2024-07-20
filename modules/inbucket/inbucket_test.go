package inbucket

import (
	"context"
	"net/smtp"
	"testing"

	"github.com/inbucket/inbucket/pkg/rest/client"
	"github.com/stretchr/testify/require"
)

func TestInbucket(t *testing.T) {
	ctx := context.Background()

	container, err := Run(ctx, "inbucket/inbucket:sha-2d409bb")
	require.NoError(t, err)

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		err := container.Terminate(ctx)
		require.NoError(t, err)
	})

	// smtpConnection {
	smtpUrl, err := container.SmtpConnection(ctx)
	// }
	require.NoError(t, err)

	// webInterface {
	webInterfaceUrl, err := container.WebInterface(ctx)
	// }
	require.NoError(t, err)
	restClient, err := client.New(webInterfaceUrl)
	require.NoError(t, err)

	headers, err := restClient.ListMailbox("to@example.org")
	require.NoError(t, err)
	require.Empty(t, headers)

	msg := []byte("To: to@example.org\r\n" +
		"Subject: Testcontainers test!\r\n" +
		"\r\n" +
		"This is a Testcontainers test.\r\n")
	err = smtp.SendMail(smtpUrl, nil, "from@example.org", []string{"to@example.org"}, msg)
	require.NoError(t, err)

	headers, err = restClient.ListMailbox("to@example.org")
	require.NoError(t, err)
	require.Len(t, headers, 1)
}
