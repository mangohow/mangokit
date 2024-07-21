
import (
	"context"
	"net/http"

	{{if .ImportSerialize}}
	"github.com/mangohow/mangokit/serialize"
	{{- end}}
	"github.com/mangohow/mangokit/tools"
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
    server.RegisterService(_{{.ServiceName}}HTTPService_serviceDesc, svc)
}

{{range .Methods}}
func _{{.ServiceName}}_{{.Name}}_HTTP_Handler(svc interface{}, middleware httpwrapper.Middleware) httpwrapper.Middleware {
	return func(ctx context.Context, req interface{}, next httpwrapper.NextHandler) error {
		{{- if ne .InputFieldLen 0}}
		in := new({{.Request}})
		err := tools.BindVar(ctx, in)
        if err != nil {
            return err
        }

		{{end}}
		handler := func(ctx context.Context, req interface{}) error {
            ctxt := tools.GinCtxFromContext(ctx)
            {{- if and (eq .InputFieldLen 0) (eq .OutputFieldLen 0)}}
                err := svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx)
            {{- else if eq .InputFieldLen 0}}
                reply, err := svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx)
            {{- else if eq .OutputFieldLen 0}}
                err := svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx, in)
            {{- else}}
                reply, err := svc.({{.ServiceName}}HTTPService).{{.Name}}(ctx, in)
            {{- end}}
            if err != nil {
                return err
            }
            {{- if ne .OutputFieldLen 0}}
            ctxt.JSON(http.StatusOK, serialize.Response{Data: reply})
            {{- else}}
            ctxt.Status(http.StatusOK)
            {{- end}}

            return nil
         }


        if middleware == nil {
            {{- if eq .InputFieldLen 0}}
            return handler(ctx, nil)
            {{- else}}
            return handler(ctx, in)
            {{- end}}
        }

        {{if eq .InputFieldLen 0}}
        return middleware(ctx, nil, handler)
        {{- else}}
        return middleware(ctx, in, handler)
        {{- end}}
	}
}
{{end}}


var _{{.ServiceName}}HTTPService_serviceDesc = &httpwrapper.ServiceDesc{
	HandlerType: (*{{.ServiceName}}HTTPService)(nil),
	Methods: []httpwrapper.MethodDesc{
	{{- range .Methods}}
		{
			Method:  "GET",
			Path:    "{{.Path}}",
			Handler: _Greeter_{{.Name}}_HTTP_Handler,
		},
	{{- end}}
	},
}
