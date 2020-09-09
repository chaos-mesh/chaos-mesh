{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "chaos-mesh.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "chaos-mesh.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "chaos-mesh.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Define the svc's name
*/}}
{{- define "chaos-mesh.svc" -}}
{{- printf "chaos-mesh-controller-manager" -}}
{{- end -}}

{{/*
Define the chaos-dashboard svc's name
*/}}
{{- define "chaos-dashboard.svc" -}}
{{- printf "chaos-dashboard" -}}
{{- end -}}

{{/*
Define the secret's name of certs
*/}}
{{- define "chaos-mesh.certs" -}}
{{- printf "chaos-mesh-webhook-certs" -}}
{{- end -}}

{{/*
Define the MutatingWebhookConfiguration's name
*/}}
{{- define "chaos-mesh.mutation" -}}
{{- printf "chaos-mesh-mutation" -}}
{{- end -}}

{{/*
Define the ValidationWebhookConfiguration's name
*/}}
{{- define "chaos-mesh.validation" -}}
{{- printf "chaos-mesh-validation" -}}
{{- end -}}

{{/*
Define the webhook's name
*/}}
{{- define "chaos-mesh.webhook" -}}
{{- printf "admission-webhook.chaos-mesh.org" -}}
{{- end -}}
