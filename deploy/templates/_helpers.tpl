{{/* vim: set filetype=mustache: */}}

{{/*
Create a default fully qualified app name.
We truncate at 53 chars because some Kubernetes name fields are limited to 63 (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- printf "%s-%s" .Chart.Name .Release.Name | trunc 53 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common metadata
*/}}
{{- define "commonMeta" }}
app: {{ template "fullname" . }}
chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
release: "{{ .Release.Name }}"
{{- end }}
