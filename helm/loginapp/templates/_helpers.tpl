{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "loginapp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "loginapp.fullname" -}}
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
{{- define "loginapp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "loginapp.labels" -}}
helm.sh/chart: {{ include "loginapp.chart" . }}
{{ include "loginapp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "loginapp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "loginapp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "loginapp.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "loginapp.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Generate redirectURL
*/}}
{{- define "loginapp.redirecturl" -}}
{{-     if and .Values.ingress.enabled (index .Values.ingress.hosts 0) -}}
{{-         if .Values.ingress.tls -}}https{{- else -}}http{{- end -}}://{{- with (index .Values.ingress.hosts 0) }}{{ .t }}{{- end -}}
{{-     else -}}
{{-         if .Values.config.tls.enabled -}}https{{- else -}}http{{- end -}}://{{ default (include "loginapp.fullname" .) }}.{{ .Release.Namespace }}.svc:{{ .Values.service.port }}
{{-     end -}}
/callback
{{- end -}}

{{/*
Generate configuration
*/}}
{{- define "loginapp.configuration" -}}
name: {{ default (include "loginapp.fullname" .) .Values.config.name }}
listen: 0.0.0.0:5555
oidc:
  client:
    id: {{ .Values.config.clientID }}
    redirectURL: {{ default (include "loginapp.redirecturl" .) .Values.config.clientRedirectURL }}
  issuer:
    rootCA: "/tls/issuer/ca.crt"
    url: {{ .Values.config.issuerURL }}
tls:
  enabled: {{ .Values.config.tls.enabled }}
  cert: /tls/server/tls.crt
  key: /tls/server/tls.key
clusters:
  {{- toYaml .Values.config.clusters | nindent 2 }}
{{- end -}}