
{{/*
webhook.apiversion is used to take care compatibility with admissionregistration.k8s.io api groups

When using this template, it requires the top-level scope
*/}}
{{- define "webhook.apiVersion" -}}
  {{- $webhookApiVersion := "v1beta1" -}}
  {{- if .Capabilities.APIVersions.Has "admissionregistration.k8s.io/v1" -}}
    {{- $webhookApiVersion = "v1" -}}
  {{- end -}}
  {{- printf "admissionregistration.k8s.io/%s" $webhookApiVersion -}}
{{- end -}}

{{/*
chaosmesh.selfSignedCABundleCertPEM is the self-signed CA to:
- sign the certification keyparing used by chaos-daemon mTLS
- sign the certification keyparing used by webhook server (if the user does not provide one)
*/}}
{{- define "chaosmesh.selfSignedCABundleCertPEM" -}}
  {{- $caKeypair := .selfSignedCAKeypair | default (genCA "chaos-mesh-ca" 1825) -}}
  {{- $_ := set . "selfSignedCAKeypair" $caKeypair -}}
  {{- $caKeypair.Cert -}}
{{- end -}}

{{/*
Get the caBundle for clients of the webhooks.
It would use .selfSignedCAKeypair as the place to store the generated CA keypair, it is actually kind of dirty work to prevent generating keypair with multiple times.

When using this template, it requires the top-level scope.

*/}}
{{- define "webhook.caBundleCertPEM" -}}
  {{- if .Values.webhook.caBundlePEM -}}
    {{- trim .Values.webhook.caBundlePEM -}}
  {{- else -}}
    {{- /* Generate ca with CN "chaos-mesh-ca" and 5 years validity duration if not exists in the current scope.*/ -}}
    {{- $caKeypair := .selfSignedCAKeypair | default (genCA "chaos-mesh-ca" 1825) -}}
    {{- $_ := set . "selfSignedCAKeypair" $caKeypair -}}
    {{- $caKeypair.Cert -}}
  {{- end -}}
{{- end -}}

{{/*
webhook.certPEM is the cert of certification used by validating/mutating admission webhook server.
Like generating CA, it would use .webhookTLSKeypair as the place to store the generated keypair, it is actually kind of dirty work to prevent generating keypair with multiple times.

When using this template, it requires the top-level scope
*/}}
{{- define "webhook.certPEM" -}}
  {{- if .Values.webhook.crtPEM -}}
    {{- trim .Values.webhook.crtPEM -}}
  {{- else -}}
    {{- /* FIXME: Duplicated codes with named template "webhook.keyPEM" because of no way to nested named template.*/ -}}
    {{- /* webhookName would be the FQDN of in-cluster service chaos-mesh.*/ -}}
    {{- $webhookName := printf "%s.%s.svc" (include "chaos-mesh.svc" .) .Release.Namespace }}
    {{- $webhookCA := required "self-signed CA keypair is requried" .selfSignedCAKeypair -}}
    {{- /* Generate cert keypair for webhook with 5 year validity duration. */ -}}
    {{- $webhookServerTLSKeypair := .webhookTLSKeypair | default (genSignedCert $webhookName nil (list $webhookName) 1825 $webhookCA) }}
    {{- $_ := set . "webhookTLSKeypair" $webhookServerTLSKeypair -}}
    {{- $webhookServerTLSKeypair.Cert -}}
  {{- end -}}
{{- end -}}

{{/*
webhook.keyPEM is the key of certification used by validating/mutating admission webhook server.
Like generating CA, it would use .webhookTLSKeypair as the place to store the generated keypair, it is actually kind of dirty work to prevent generating keypair with multiple times.

When using this template, it requires the top-level scope
*/}}
{{- define "webhook.keyPEM" -}}
  {{- if .Values.webhook.keyPEM -}}
    {{ trim .Values.webhook.keyPEM }}
  {{- else -}}
    {{- /* FIXME: Duplicated codes with named template "webhook.keyPEM" because of no way to nested named template.*/ -}}
    {{- /* webhookName would be the FQDN of in-cluster service chaos-mesh.*/ -}}
    {{- $webhookName := printf "%s.%s.svc" (include "chaos-mesh.svc" .) .Release.Namespace -}}
    {{- $webhookCA := required "self-signed CA keypair is requried" .selfSignedCAKeypair -}}
    {{- /* Generate cert key pair for webhook with 5 year validity duration. */ -}}
    {{- $webhookServerTLSKeypair := .webhookTLSKeypair | default (genSignedCert $webhookName nil (list $webhookName) 1825 $webhookCA) -}}
    {{- $_ := set . "webhookTLSKeypair" $webhookServerTLSKeypair -}}
    {{- $webhookServerTLSKeypair.Key -}}
  {{- end -}}
{{- end -}}

{{/*
chaosDaemon.server.certPEM is the certification used by chaos daemon server for mTLS.
Like generating CA, it would use .chaosDaemonServerTLSKeypair as the place to store the generated keypair,
it is actually kind of dirty work to prevent generating keypair with multiple times.

When using this template, it requires the top-level scope.
*/}}
{{- define "chaosDaemon.server.certPEM" -}}
    {{- $ca := required "self-signed CA keypair is requried" .selfSignedCAKeypair -}}
    {{- /* Generate cert keypair with CN "chaos-daemon.chaos-mesh.org" and 5 years validity duration if not exists in the current scope.*/ -}}
    {{- $chaosDaemonServerTLSKeypair := .chaosDaemonServerTLSKeypair | default (genSignedCert "chaos-daemon.chaos-mesh.org" nil (list "localhost" "chaos-daemon.chaos-mesh.org") 1825 $ca) -}}
    {{- $_ := set . "chaosDaemonServerTLSKeypair" $chaosDaemonServerTLSKeypair -}}
    {{- $chaosDaemonServerTLSKeypair.Cert -}}
{{- end -}}

{{/*
chaosDaemon.server.keyPEM is the key used by chaos daemon server for mTLS.
Like generating CA, it would use .chaosDaemonServerTLSKeypair as the place to store the generated keypair,
it is actually kind of dirty work to prevent generating keypair with multiple times.

When using this template, it requires the top-level scope.
*/}}
{{- define "chaosDaemon.server.keyPEM" -}}
    {{- $ca := required "self-signed CA keypair is requried" .selfSignedCAKeypair -}}
    {{- /* Generate cert keypair with CN "chaos-daemon.chaos-mesh.org" and 5 years validity duration if not exists in the current scope.*/ -}}
    {{- $chaosDaemonServerTLSKeypair := .chaosDaemonServerTLSKeypair | default (genSignedCert "chaos-daemon.chaos-mesh.org" nil (list "localhost" "chaos-daemon.chaos-mesh.org") 1825 $ca) -}}
    {{- $_ := set . "chaosDaemonServerTLSKeypair" $chaosDaemonServerTLSKeypair -}}
    {{- $chaosDaemonServerTLSKeypair.Key -}}
{{- end -}}

{{/*
chaosDaemon.client.certPEM is the certification used by controller-manager (as the client of chaos-daemon server) for mTLS.
Like generating CA, it would use .chaosDaemonClientTLSKeypair as the place to store the generated keypair,
it is actually kind of dirty work to prevent generating keypair with multiple times.

When using this template, it requires the top-level scope.
*/}}
{{- define "chaosDaemon.client.certPEM" -}}
    {{- $ca := required "self-signed CA keypair is requried" .selfSignedCAKeypair -}}
    {{- /* Generate cert keypair with CN "controller-manager.chaos-mesh.org" and 5 years validity duration if not exists in the current scope.*/ -}}
    {{- $chaosDaemonClientTLSKeypair := .chaosDaemonClientTLSKeypair | default (genSignedCert "controller-manager.chaos-mesh.org" nil (list "localhost" "controller-manager.chaos-mesh.org") 1825 $ca) -}}
    {{- $_ := set . "chaosDaemonClientTLSKeypair" $chaosDaemonClientTLSKeypair -}}
    {{- $chaosDaemonClientTLSKeypair.Cert -}}
{{- end -}}

{{/*
chaosDaemon.client.keyPEM is the key used by controller-manager (as the client of chaos-daemon server) for mTLS.
Like generating CA, it would use .chaosDaemonClientTLSKeypair as the place to store the generated keypair,
it is actually kind of dirty work to prevent generating keypair with multiple times.

When using this template, it requires the top-level scope.
*/}}
{{- define "chaosDaemon.client.keyPEM" -}}
    {{- $ca := required "self-signed CA keypair is requried" .selfSignedCAKeypair -}}
    {{- /* Generate cert keypair with CN "controller-manager.chaos-mesh.org" and 5 years validity duration if not exists in the current scope.*/ -}}
    {{- $chaosDaemonClientTLSKeypair := .chaosDaemonClientTLSKeypair | default (genSignedCert "controller-manager.chaos-mesh.org" nil (list "localhost" "controller-manager.chaos-mesh.org") 1825 $ca) -}}
    {{- $_ := set . "chaosDaemonClientTLSKeypair" $chaosDaemonClientTLSKeypair -}}
    {{- $chaosDaemonClientTLSKeypair.Key -}}
{{- end -}}
