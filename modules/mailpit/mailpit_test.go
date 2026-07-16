package mailpit_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mailpit"
)

func TestMailpit(t *testing.T) {
	ctx := context.Background()

	ctr, err := mailpit.Run(ctx, "axllent/mailpit:v1.20")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// smtpEndpoint {
	smtpEndpoint, err := ctr.SMTPEndpoint(ctx)
	// }
	require.NoError(t, err)
	require.NotEmpty(t, smtpEndpoint)

	// httpURL {
	httpURL, err := ctr.HTTPURL(ctx)
	// }
	require.NoError(t, err)
	require.NotEmpty(t, httpURL)
}

func TestMailpitWithSMTPAuth(t *testing.T) {
	ctx := context.Background()

	// withSMTPAuth {
	ctr, err := mailpit.Run(ctx, "axllent/mailpit:v1.20",
		mailpit.WithSMTPAuth("user", "password"),
	)
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	smtpEndpoint, err := ctr.SMTPEndpoint(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, smtpEndpoint)
}

func TestMailpitWithMessageLimit(t *testing.T) {
	ctx := context.Background()

	// withMessageLimit {
	ctr, err := mailpit.Run(ctx, "axllent/mailpit:v1.20",
		mailpit.WithMessageLimit(100),
	)
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	httpURL, err := ctr.HTTPURL(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, httpURL)
}
