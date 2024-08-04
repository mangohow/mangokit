
{{ if .GenDesc}}
const (
	{{- range .Errors}}
	{{- if ne .Desc ""}}
	Desc_{{ .Name }} = "{{ .Desc }}"
	{{- end }}
	{{- end }}
)
{{ end }}
{{ if .GenDesc}}
var (
	{{- range .Errors }}
	{{- if ne .Desc "" }}
	Error{{ .Name }} = errors.New(int32({{ .EnumName }}_{{ .Name }}), {{ .HTTPStatus }}, "{{ .Name }}", Desc_{{ .Name }})
	{{- end}}
	{{- end }}
)
{{ end }}

{{ range .Errors}}
{{ if ne .Comment ""}}{{ .Comment }}{{ end -}}
func NewError{{ .CamelName }}(format string, args ...interface{}) errors.Error {
	return errors.New(int32({{ .EnumName }}_{{ .Name }}), {{ .HTTPStatus }}, {{ .EnumName }}_{{ .Name }}.String(), fmt.Sprintf(format, args...))
}

{{ end }}
