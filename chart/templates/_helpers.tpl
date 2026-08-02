{{- define "gexec.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "gexec.fullname" -}}
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

{{- define "gexec.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "gexec.labels" -}}
helm.sh/chart: "{{ include "gexec.chart" . }}"
app.kubernetes.io/name: "gexec"
app.kubernetes.io/instance: "{{ .Release.Name }}"
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
{{- with .Values.labels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "gexec.server.labels" -}}
{{- include "gexec.labels" . }}
app.kubernetes.io/component: server
{{- end -}}

{{- define "gexec.runner.labels" -}}
{{- include "gexec.labels" . }}
app.kubernetes.io/component: runner
{{- end -}}

{{- define "gexec.cleanup.labels" -}}
{{- include "gexec.labels" . }}
app.kubernetes.io/component: cleanup
{{- end -}}

{{- define "gexec.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default "gexec" .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{- define "gexec.server.selectorLabels" -}}
app.kubernetes.io/name: gexec
app.kubernetes.io/component: server
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "gexec.runner.selectorLabels" -}}
app.kubernetes.io/name: gexec
app.kubernetes.io/component: runner
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "gexec.database.secretName" -}}
{{- .Values.config.database.existingSecret | default (printf "%s-database" (include "gexec.fullname" .)) -}}
{{- end -}}

{{- define "gexec.token.secretName" -}}
{{- .Values.config.token.existingSecret | default (printf "%s-token" (include "gexec.fullname" .)) -}}
{{- end -}}

{{- define "gexec.admin.secretName" -}}
{{- .Values.config.admin.existingSecret | default (printf "%s-admin" (include "gexec.fullname" .)) -}}
{{- end -}}

{{- define "gexec.scim.secretName" -}}
{{- .Values.config.scim.existingSecret | default (printf "%s-scim" (include "gexec.fullname" .)) -}}
{{- end -}}

{{- define "gexec.runner.secretName" -}}
{{- .Values.config.runner.existingSecret | default (printf "%s-runner" (include "gexec.fullname" .)) -}}
{{- end -}}

{{- define "gexec.encrypt.secretName" -}}
{{- .Values.config.encrypt.existingSecret | default (printf "%s-encrypt" (include "gexec.fullname" .)) -}}
{{- end -}}

{{- define "gexec.shared.environment" -}}
- name: GEXEC_LOG_LEVEL
  value: "{{ .Values.config.log.level }}"
- name: GEXEC_LOG_PRETTY
  value: "false"
- name: GEXEC_LOG_COLOR
  value: "false"
{{- if eq .Values.config.database.driver "sqlite3" }}
- name: GEXEC_DATABASE_DRIVER
  value: "{{ .Values.config.database.driver }}"
- name: GEXEC_DATABASE_NAME
  value: "{{ .Values.config.database.name }}"
{{- else }}
- name: GEXEC_DATABASE_DRIVER
  value: "{{ .Values.config.database.driver }}"
- name: GEXEC_DATABASE_ADDRESS
  value: "{{ .Values.config.database.address }}"
- name: GEXEC_DATABASE_PORT
  value: "{{ .Values.config.database.port }}"
- name: GEXEC_DATABASE_NAME
  value: "{{ .Values.config.database.name }}"
- name: GEXEC_DATABASE_USERNAME
  value: "{{ .Values.config.database.username }}"
- name: GEXEC_DATABASE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: "{{ include "gexec.database.secretName" . }}"
      key: "{{ .Values.config.database.passwordKey }}"
{{- end }}
{{- if eq .Values.config.upload.driver "file" }}
- name: GEXEC_UPLOAD_DRIVER
  value: "{{ .Values.config.upload.driver }}"
- name: GEXEC_UPLOAD_PATH
  value: "{{ .Values.config.upload.path }}"
- name: GEXEC_UPLOAD_PERMS
  value: "{{ .Values.config.upload.perms }}"
{{- end }}
{{- if eq .Values.config.upload.driver "s3" }}
- name: GEXEC_UPLOAD_DRIVER
  value: "{{ .Values.config.upload.driver }}"
- name: GEXEC_UPLOAD_ENDPOINT
  value: "{{ .Values.config.upload.endpoint }}"
- name: GEXEC_UPLOAD_BUCKET
  value: "{{ .Values.config.upload.bucket }}"
- name: GEXEC_UPLOAD_REGION
  value: "{{ .Values.config.upload.region }}"
- name: GEXEC_UPLOAD_PATHSTYLE
  value: "{{ .Values.config.upload.pathstyle }}"
- name: GEXEC_UPLOAD_PATH
  value: "{{ .Values.config.upload.path }}"
- name: GEXEC_UPLOAD_ACCESS
  valueFrom:
    secretKeyRef:
      name: "{{ include "gexec.upload.secretName" . }}"
      key: "{{ .Values.config.upload.accessKey }}"
- name: GEXEC_UPLOAD_SECRET
  valueFrom:
    secretKeyRef:
      name: "{{ include "gexec.upload.secretName" . }}"
      key: "{{ .Values.config.upload.secretKey }}"
- name: GEXEC_UPLOAD_PROXY
  value: "{{ .Values.config.upload.proxy }}"
- name: GEXEC_UPLOAD_PERMS
  value: "{{ .Values.config.upload.perms }}"
{{- end }}
- name: GEXEC_ENCRYPT_PASSPHRASE
  valueFrom:
    secretKeyRef:
      name: "{{ include "gexec.encrypt.secretName" . }}"
      key: "{{ .Values.config.encrypt.passphraseKey }}"
{{- end -}}

{{- define "gexec.server.environment" -}}
{{- include "gexec.shared.environment" . }}
- name: GEXEC_SERVER_HOST
  value: "{{ .Values.config.server.host }}"
- name: GEXEC_SERVER_ROOT
  value: "{{ .Values.config.server.root }}"
- name: GEXEC_SERVER_DOCS
  value: "{{ .Values.config.server.docs }}"
- name: GEXEC_TOKEN_EXPIRE
  value: "{{ .Values.config.token.expire }}"
- name: GEXEC_TOKEN_SECRET
  valueFrom:
    secretKeyRef:
      name: "{{ include "gexec.token.secretName" . }}"
      key: "{{ .Values.config.token.secretKey }}"
- name: GEXEC_ADMIN_CREATE
  value: "{{ .Values.config.admin.create }}"
{{- if .Values.config.admin.create }}
- name: GEXEC_ADMIN_USERNAME
  valueFrom:
    secretKeyRef:
      name: "{{ include "gexec.admin.secretName" . }}"
      key: "{{ .Values.config.admin.usernameKey }}"
- name: GEXEC_ADMIN_PASSWORD
  valueFrom:
    secretKeyRef:
      name: "{{ include "gexec.admin.secretName" . }}"
      key: "{{ .Values.config.admin.passwordKey }}"
- name: GEXEC_ADMIN_EMAIL
  valueFrom:
    secretKeyRef:
      name: "{{ include "gexec.admin.secretName" . }}"
      key: "{{ .Values.config.admin.emailKey }}"
{{- end }}
- name: GEXEC_SCIM_ENABLED
  value: "{{ .Values.config.scim.enabled }}"
{{- if .Values.config.scim.enabled }}
- name: GEXEC_SCIM_TOKEN
  valueFrom:
    secretKeyRef:
      name: "{{ include "gexec.scim.secretName" . }}"
      key: "{{ .Values.config.scim.tokenKey }}"
{{- end }}
- name: GEXEC_AUTH_CONFIG
  value: "/etc/gexec/auth/config.yaml"
{{- end -}}

{{- define "gexec.runner.environment" -}}
{{- include "gexec.shared.environment" . }}
- name: GEXEC_RUNNER_SERVER
  value: "{{ .Values.config.runner.server }}"
- name: GEXEC_RUNNER_TOKEN
  valueFrom:
    secretKeyRef:
      name: "{{ include "gexec.runner.secretName" . }}"
      key: "{{ .Values.config.runner.tokenKey }}"
{{- end -}}

{{- define "gexec.cleanup.environment" -}}
{{- include "gexec.shared.environment" . }}
{{- end -}}
