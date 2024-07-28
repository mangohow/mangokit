
import (
	"context"

	"github.com/mangohow/mangokit/transport/http"
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

func Register{{.ServiceName}}HTTPService(server *http.Server, svc {{.ServiceName}}HTTPService) {
    server.RegisterService(_{{.ServiceName}}HTTPService_serviceDesc, svc)
}

{{range .Methods}}
func _{{.ServiceName}}_{{.Name}}_HTTP_Handler(svc interface{}, ctx context.Context, dec func(interface{}) error, middleware http.Middleware) (interface{}, error) {
    {{- if ne .InputFieldLen 0}}
    in := new({{.Request}})
    err := dec(in)
    if err != nil {
        return nil, err
    }
    {{- end}}
    {{- if ne .InputFieldLen 0}}
    {{end}}
    if middleware == nil {
    {{- if and (eq .InputFieldLen 0) (eq .OutputFieldLen 0)}}
        return nil, svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx)
    {{- else if eq .InputFieldLen 0}}
        return svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx)
    {{- else if eq .OutputFieldLen 0}}
        return nil, svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx, in)
    {{- else}}
        return svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx, in)
    {{- end}}
    }

    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
    {{- if and (eq .InputFieldLen 0) (eq .OutputFieldLen 0)}}
        return nil, svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx)
    {{- else if eq .InputFieldLen 0}}
        return svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx)
    {{- else if eq .OutputFieldLen 0}}
        return nil, svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx, in)
    {{- else}}
        return svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx, in)
    {{- end}}
    }

    {{if eq .InputFieldLen 0 }}
    return middleware(ctx, nil, handler)
    {{else}}
    return middleware(ctx, in, handler)
    {{end}}
}
{{end}}


var _{{.ServiceName}}HTTPService_serviceDesc = &http.ServiceDesc{
	HandlerType: (*{{.ServiceName}}HTTPService)(nil),
	Methods: []http.MethodDesc{
	{{- range .Methods}}
		{
			Method:  "GET",
			Path:    "{{.Path}}",
			Handler: _Greeter_{{.Name}}_HTTP_Handler,
		},
	{{- end}}
	},
}
