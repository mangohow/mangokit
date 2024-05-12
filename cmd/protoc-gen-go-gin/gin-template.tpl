
import (
	"context"
	"net/http"

	{{if .ImportSerialize}}
	"github.com/mangohow/mangokit/serialize"
	{{- end}}
	"github.com/mangohow/mangokit/transport/httpwrapper"
)

{{- if ne .Comment ""}}
{{.Comment}}
{{- end}}
type {{.ServiceName}}HTTPService interface {
{{- range .Methods}}
	{{- if ne .Comment ""}}
	{{.Comment}}
	{{- end}}
	{{- if and (eq .InputFieldLen 0) (eq .OutputFieldLen 0)}}
        {{.Name}}(context.Context) error
    {{- else if eq .InputFieldLen 0}}
        {{.Name}}(context.Context) (*{{.Reply}}, error)
    {{- else if eq .OutputFieldLen 0}}
        {{.Name}}(context.Context, *{{.Request}}) error
    {{- else}}
        {{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
    {{- end}}
{{- end}}
}

func Register{{.ServiceName}}HTTPService(server *httpwrapper.Server, svc {{.ServiceName}}HTTPService) {
	{{- range .Methods}}
	server.{{.HttpMethod}}("{{.Path}}", _{{.ServiceName}}_{{.Name}}_HTTP_Handler(svc))
	{{- end}}
}

{{range .Methods}}
func _{{.ServiceName}}_{{.Name}}_HTTP_Handler(svc {{.ServiceName}}HTTPService) httpwrapper.HandlerFunc {
	return func(ctx *httpwrapper.Context) error {
		{{- if ne .InputFieldLen 0}}
		in := new({{.Request}})
		if err := ctx.BindRequest(in); err != nil {
			return err
		}

		{{- end}}
		value := context.WithValue(context.Background(), "gin-ctx", ctx)
		{{- if and (eq .InputFieldLen 0) (eq .OutputFieldLen 0)}}
            err := svc.{{.Name}}(value)
        {{- else if eq .InputFieldLen 0}}
            reply, err := svc.{{.Name}}(value)
        {{- else if eq .OutputFieldLen 0}}
            err := svc.{{.Name}}(value, in)
        {{- else}}
            reply, err := svc.{{.Name}}(value, in)
        {{- end}}
		if err != nil {
			return err
		}
		{{- if ne .OutputFieldLen 0}}
		ctx.JSON(http.StatusOK, serialize.Response{Data: reply})
		{{- else}}
		ctx.Status(http.StatusOK)
		{{- end}}

		return nil
	}
}
{{end}}