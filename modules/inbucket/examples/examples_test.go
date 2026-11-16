package inbucket_test

import (
	"context"
	"fmt"
	"log"
	"net/smtp"

	"github.com/inbucket/inbucket/pkg/rest/client"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/inbucket"
)

func ExampleRun_mailbox() {
	// mailboxDelivery {
	ctx := context.Background()
	ctr, err := inbucket.Run(ctx, "inbucket/inbucket:sha-2d409bb")
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	smtpURL, err := ctr.SmtpConnection(ctx)
	if err != nil {
		log.Printf("failed to get SMTP connection: %s", err)
		return
	}
	webURL, err := ctr.WebInterface(ctx)
	if err != nil {
		log.Printf("failed to get web interface: %s", err)
		return
	}
	restClient, err := client.New(webURL)
	if err != nil {
		log.Printf("failed to create Inbucket client: %s", err)
		return
	}

	headers, err := restClient.ListMailbox("to@example.org")
	if err != nil {
		log.Printf("failed to list empty mailbox: %s", err)
		return
	}
	fmt.Printf("Messages before delivery: %d\n", len(headers))

	msg := []byte("To: to@example.org\r\n" +
		"Subject: Testcontainers test!\r\n" +
		"\r\n" +
		"This is a Testcontainers test.\r\n")
	if err := smtp.SendMail(smtpURL, nil, "from@example.org", []string{"to@example.org"}, msg); err != nil {
		log.Printf("failed to send message: %s", err)
		return
	}
	headers, err = restClient.ListMailbox("to@example.org")
	if err != nil {
		log.Printf("failed to list mailbox after delivery: %s", err)
		return
	}
	fmt.Printf("Messages after delivery: %d\n", len(headers))
	// }

	// Output:
	// Messages before delivery: 0
	// Messages after delivery: 1
}
