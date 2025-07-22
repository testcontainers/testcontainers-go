package solace

import (
	"fmt"
)

// generateCLIScript generates a Solace CLI script for configuring VPN, users, queues, and topic subscriptions.
// Reference: https://docs.solace.com/Admin-Ref/CLI-Reference/VMR_CLI_Commands.html
func generateCLIScript(settings options) string {
	if len(settings.queues) == 0 {
		return ""
	}

	script := `enable
configure
`

	// Create VPN if not default
	if settings.vpn != defaultVpn {
		script += generateVPNConfig(settings.vpn)
	}

	// Configure username and password
	script += generateUserConfig(settings.username, settings.password, settings.vpn)

	// Configure VPN Basic authentication
	script += generateVPNAuth(settings.vpn)

	// Configure queues and topic subscriptions
	script += generateQueueConfig(settings.queues, settings.vpn)

	return script
}

// generateVPNConfig creates the CLI commands for setting up a custom VPN
func generateVPNConfig(vpn string) string {
	return fmt.Sprintf(`create message-vpn %s
	no shutdown
	exit
	client-profile default message-vpn %s
		message-spool
			allow-guaranteed-message-send
			allow-guaranteed-message-receive
			allow-guaranteed-endpoint-create
			allow-guaranteed-endpoint-create-durability all
			exit
		exit
	message-spool message-vpn %s
		max-spool-usage 60000
		exit
`, vpn, vpn, vpn)
}

// generateUserConfig creates the CLI commands for setting up user authentication
func generateUserConfig(username, password, vpn string) string {
	return fmt.Sprintf(`create client-username %s message-vpn %s
	password %s
	no shutdown
	exit
`, username, vpn, password)
}

// generateVPNAuth creates the CLI commands for setting up VPN authentication
func generateVPNAuth(vpn string) string {
	return fmt.Sprintf(`message-vpn %s
	authentication basic auth-type internal
	no shutdown
	end
`, vpn)
}

// generateQueueConfig creates the CLI commands for setting up queues and their topic subscriptions
func generateQueueConfig(queues map[string][]string, vpn string) string {
	script := fmt.Sprintf(`configure
message-spool message-vpn %s
`, vpn)

	for queue, topics := range queues {
		// Create the queue first
		script += fmt.Sprintf(`	create queue %s
		access-type exclusive
		permission all consume
		no shutdown
		exit
`, queue)

		// Add topic subscriptions to the queue
		for _, topic := range topics {
			script += fmt.Sprintf(`	queue %s
		subscription topic %s
		exit
`, queue, topic)
		}
	}

	script += `	exit
exit
`
	return script
}
