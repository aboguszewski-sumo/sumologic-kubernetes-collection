{{- define "sumologic.metadata.name.pvc-cleaner" -}}
{{- template "sumologic.fullname" . }}-pvc-cleaner
{{- end -}}

{{- define "sumologic.metadata.name.pvc-cleaner.logs" -}}
{{- template "sumologic.metadata.name.pvc-cleaner" . }}-logs
{{- end -}}

{{- define "sumologic.metadata.name.pvc-cleaner.metrics" -}}
{{- template "sumologic.metadata.name.pvc-cleaner" . }}-metrics
{{- end -}}

{{- define "sumologic.labels.app.pvc-cleaner" -}}
pvc-cleaner
{{- end -}}

{{- define "sumologic.labels.app.pvc-cleaner.logs" -}}
{{- template "sumologic.labels.app.pvc-cleaner" . }}-logs
{{- end -}}

{{- define "sumologic.labels.app.pvc-cleaner.metrics" -}}
{{- template "sumologic.labels.app.pvc-cleaner" . }}-metrics
{{- end -}}

{{- define "sumologic.metadata.name.pvc-cleaner.roles.serviceaccount" -}}
{{- template "sumologic.fullname" . }}-pvc-cleaner
{{- end -}}

{{- define "sumologic.labels.app.pvc-cleaner.roles.serviceaccount" -}}
{{- template "sumologic.fullname" . }}-pvc-cleaner
{{- end -}}