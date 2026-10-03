package papercutsmtp_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/papercutsmtp"
)

func TestPapercutSMTP(t *testing.T) {
	ctx := context.Background()

	ctr, err := papercutsmtp.Run(ctx, "changemakerstudiosus/papercut-smtp:latest")
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
	require.Contains(t, httpURL, "http://")
}
