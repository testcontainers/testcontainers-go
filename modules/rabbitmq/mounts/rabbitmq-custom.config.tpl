%% Injected by testcontainers
[
{rabbit, [
{{- if .SSLSettings }}
   {ssl_listeners, [5671]},
   {ssl_options, [{cacertfile,"{{ .SSLSettings.CACertFile }}"},
                  {certfile,"{{ .SSLSettings.CertFile }}"},
                  {keyfile,"{{ .SSLSettings.KeyFile }}"},
                  {depth, {{ .SSLSettings.VerificationDepth }}},
                  {verify, {{ .SSLSettings.VerificationMode }}},
                  {fail_if_no_peer_cert, {{ .SSLSettings.FailIfNoCert }}}]}
{{- end }}
 ]}
].
