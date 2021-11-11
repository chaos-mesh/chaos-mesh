{{- define "type" -}}
{{- $type := . -}}
{{- if asciidocShouldRenderType $type -}}

[id="{{ asciidocTypeID $type | asciidocRenderAnchorID }}"]
==== {{ $type.Name  }} {{ if $type.IsAlias }}({{ asciidocRenderTypeLink $type.UnderlyingType  }}) {{ end }}

{{ $type.Doc }}

{{ if $type.References -}}
.Appears In:
****
{{- range $type.SortedReferences }}
- {{ asciidocRenderTypeLink . }}
{{- end }}
****
{{- end }}

{{ if $type.Members -}}
[cols="25a,75a", options="header"]
|===
| Field | Description
{{ if $type.GVK -}}
| *`apiVersion`* __string__ | `{{ $type.GVK.Group }}/{{ $type.GVK.Version }}`
| *`kind`* __string__ | `{{ $type.GVK.Kind }}`
{{ end -}}

{{ range $type.Members -}}
| *`{{ .Name  }}`* __{{ asciidocRenderType .Type }}__ | {{ template "type_members" . }}
{{ end -}}
|===
{{ end -}}

{{- end -}}
{{- end -}}
