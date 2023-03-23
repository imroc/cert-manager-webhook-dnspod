{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "dnspod-webhook.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "dnspod-webhook.fullname" -}}
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
{{- define "dnspod-webhook.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "dnspod-webhook.selfSignedIssuer" -}}
{{ printf "%s-selfsign" (include "dnspod-webhook.fullname" .) }}
{{- end -}}

{{- define "dnspod-webhook.rootCAIssuer" -}}
{{ printf "%s-ca" (include "dnspod-webhook.fullname" .) }}
{{- end -}}

{{- define "dnspod-webhook.rootCACertificate" -}}
{{ printf "%s-ca" (include "dnspod-webhook.fullname" .) }}
{{- end -}}

{{- define "dnspod-webhook.servingCertificate" -}}
{{ printf "%s-webhook-tls" (include "dnspod-webhook.fullname" .) }}
{{- end -}}

{{- define "dnspod-webhook.clusterIssuer" -}}
{{- if .Values.clusterIssuer.name -}}
{{ .Values.clusterIssuer.name }}
{{- else -}}
{{ printf "%s-cluster-issuer" (include "dnspod-webhook.fullname" .) }}
{{- end -}}
{{- end -}}


{{- define "dnspod-webhook.namespace" -}}
    {{ .Values.namespace | default .Release.Namespace }}
{{- end -}}