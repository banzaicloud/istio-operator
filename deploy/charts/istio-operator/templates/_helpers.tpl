{{/*
Expand the name of the chart.
*/}}
{{- define "istio-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "istio-operator.fullname" -}}
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
{{- define "istio-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "istio-operator.labels" -}}
app: {{ include "istio-operator.fullname" . }}
helm.sh/chart: {{ include "istio-operator.chart" . }}
{{ include "istio-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | replace "+" "_" | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: {{ include "istio-operator.name" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "istio-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "istio-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Authproxy labels
*/}}
{{- define "istio-operator.authProxyLabels" -}}
{{ include "istio-operator.labels" . }}
app.kubernetes.io/component: authproxy
{{- end }}

{{/*
Operator labels
*/}}
{{- define "istio-operator.operatorLabels" -}}
{{ include "istio-operator.labels" . }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Operator selector labels
*/}}
{{- define "istio-operator.operatorSelectorLabels" -}}
{{ include "istio-operator.selectorLabels" . }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Authproxy resource name
*/}}
{{- define "istio-operator.authProxyName" -}}
{{ include "istio-operator.fullname" . }}-authproxy
{{- end }}
