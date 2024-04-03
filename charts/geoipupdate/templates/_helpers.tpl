{{/*
Copyright VMware, Inc.
SPDX-License-Identifier: APACHE-2.0
*/}}

{{/*
Return the proper geoipupdate image name
*/}}
{{- define "geoipupdate.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.geoipupdate.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "geoipupdate.imagePullSecrets" -}}
{{- include "common.images.renderPullSecrets" (dict "images" (list .Values.geoipupdate.image) "context" $) -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "geoipupdate.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "common.names.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}
