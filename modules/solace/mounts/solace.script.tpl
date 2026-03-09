enable
configure

{{- if ne .VPN "default" }}
create message-vpn {{ .VPN }}
	no shutdown
	exit
	client-profile default message-vpn {{ .VPN }}
		message-spool
			allow-guaranteed-message-send
			allow-guaranteed-message-receive
			allow-guaranteed-endpoint-create
			allow-guaranteed-endpoint-create-durability all
			exit
		exit
	message-spool message-vpn {{ .VPN }}
		max-spool-usage 60000
		exit
{{- end }}

create client-username {{ .Username }} message-vpn {{ .VPN }}
	password {{ .Password }}
	no shutdown
	exit

message-vpn {{ .VPN }}
	authentication basic auth-type internal
	no shutdown
	end

configure
message-spool message-vpn {{ .VPN }}
{{- range $queue, $topics := .Queues }}
	create queue {{ $queue }}
		access-type exclusive
		permission all consume
		no shutdown
		exit
{{- range $topics }}
	queue {{ $queue }}
		subscription topic {{ . }}
		exit
{{- end }}
{{- end }}
	exit
exit