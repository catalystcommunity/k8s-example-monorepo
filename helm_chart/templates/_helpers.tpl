{{/*
Expand the name of the chart.
*/}}
{{- define "auth.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 59 | trimSuffix "-" }}{{ "auth" }}
{{- end }}
{{- define "app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 59 | trimSuffix "-" }}{{ "app" }}
{{- end }}
{{- define "web.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 59 | trimSuffix "-" }}{{ "web" }}
{{- end }}


{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "auth.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 59 | trimSuffix "-" }}{{ "auth" }}
{{- end }}
{{- define "app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 59 | trimSuffix "-" }}{{ "app" }}
{{- end }}
{{- define "web.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 59| trimSuffix "-" }}{{ "web" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "auth.labels" -}}
helm.sh/chart: {{ include "auth.chart" . }}
{{ include "auth.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
{{- define "app.labels" -}}
helm.sh/chart: {{ include "app.chart" . }}
{{ include "app.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
{{- define "web.labels" -}}
helm.sh/chart: {{ include "web.chart" . }}
{{ include "web.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "auth.selectorLabels" -}}
app.kubernetes.io/name: {{ include "auth.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
{{- define "app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
{{- define "web.selectorLabels" -}}
app.kubernetes.io/name: {{ include "web.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "auth.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "auth.name" .) .Values.serviceAccount.name }}{{ "auth" }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}{{ "auth" }}
{{- end }}
{{- end }}
{{- define "app.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "app.name" .) .Values.serviceAccount.name }}{{ "app" }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}{{ "app" }}
{{- end }}
{{- end }}
{{- define "web.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "web.name" .) .Values.serviceAccount.name }}{{ "web" }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}{{ "web" }}
{{- end }}
{{- end }}

{{/*
Create the name of the secret. Default to the service name, allow overriding.
*/}}
{{- define "auth.secretName" -}}
{{- default (include "auth.name" .) .Values.secretName }}
{{- end }}
{{- define "app.secretName" -}}
{{- default (include "app.name" .) .Values.secretName }}
{{- end }}
{{- define "web.secretName" -}}
{{- default (include "web.name" .) .Values.secretName }}
{{- end }}

{{/*
Create migrate job name. For hosted environments, default to the app version.
This prevents problems with the job's image being immutable.
*/}}
{{- define "app.migrationJobName" -}}
{{- if .Values.migrations.setJobNameAsTimestamp }}
{{- printf "%s-migrate-%s" .Values.migrations.jobName (now | date "20060102150405") }}
{{- else }}
{{- printf "%s-migrate-%s" .Values.migrations.jobName .Chart.AppVersion }}
{{- end }}
{{- end }}
