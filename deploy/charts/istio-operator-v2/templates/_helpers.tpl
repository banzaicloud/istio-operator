{{/*
Expand the name of the chart.
*/}}
{{- define "istio-operator-v2.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "istio-operator-v2.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "istio-operator-v2.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "istio-operator-v2.labels" -}}
app: {{ include "istio-operator-v2.fullname" . }}
helm.sh/chart: {{ include "istio-operator-v2.chart" . }}
{{ include "istio-operator-v2.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | replace "+" "_" | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: {{ include "istio-operator-v2.name" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "istio-operator-v2.selectorLabels" -}}
app.kubernetes.io/name: {{ include "istio-operator-v2.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Authproxy labels
*/}}
{{- define "istio-operator-v2.authProxyLabels" -}}
{{ include "istio-operator-v2.labels" . }}
app.kubernetes.io/component: authproxy
{{- end }}

{{/*
Operator labels
*/}}
{{- define "istio-operator-v2.operatorLabels" -}}
{{ include "istio-operator-v2.labels" . }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Operator selector labels
*/}}
{{- define "istio-operator-v2.operatorSelectorLabels" -}}
{{ include "istio-operator-v2.selectorLabels" . }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Authproxy resource name
*/}}
{{- define "istio-operator-v2.authProxyName" -}}
{{ include "istio-operator-v2.fullname" . }}-authproxy
{{- end }}
