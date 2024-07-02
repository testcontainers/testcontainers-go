package inbucket

import (
	"context"
	"net/smtp"
	"testing"

	"github.com/inbucket/inbucket/pkg/rest/client"
)

func TestInbucket(t *testing.T) {
	ctx := context.Background()

	container, err := Run(ctx, "inbucket/inbucket:sha-2d409bb")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// smtpConnection {
	smtpUrl, err := container.SmtpConnection(ctx)
	// }
	if err != nil {
		t.Fatal(err)
	}

	// webInterface {
	webInterfaceUrl, err := container.WebInterface(ctx)
	// }
	if err != nil {
		t.Fatal(err)
	}

	restClient, err := client.New(webInterfaceUrl)
	if err != nil {
		t.Fatal(err)
	}

	headers, err := restClient.ListMailbox("to@example.org")
	if err != nil {
		t.Fatal(err)
	}
	if len(headers) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(headers))
	}

	msg := []byte("To: to@example.org\r\n" +
		"Subject: Testcontainers test!\r\n" +
		"\r\n" +
		"This is a Testcontainers test.\r\n")
	if err = smtp.SendMail(smtpUrl, nil, "from@example.org", []string{"to@example.org"}, msg); err != nil {
		t.Fatal(err)
	}

	headers, err = restClient.ListMailbox("to@example.org")
	if err != nil {
		t.Fatal(err)
	}

	if len(headers) != 1 {
		t.Fatalf("expected 1 message, got %d", len(headers))
	}
	// perform assertions
}
