
import (
	"context"
	"net/http"

	"github.com/mangohow/mangokit/serialize"
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
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceName}}HTTPService(server *httpwrapper.Server, svc GreeterHTTPService) {
	{{- range .Methods}}
	server.{{.HttpMethod}}("{{.Path}}", _{{.ServiceName}}_{{.Name}}_HTTP_Handler(svc))
	{{- end}}
}

{{range .Methods}}
func _{{.ServiceName}}_{{.Name}}_HTTP_Handler(svc {{.ServiceName}}HTTPService) httpwrapper.HandlerFunc {
	return func(ctx *httpwrapper.Context) error {
		in := new({{.Request}})
		if err := ctx.BindRequest(in); err != nil {
			return err
		}

		value := context.WithValue(context.Background(), "gin-ctx", ctx)
		reply, err := svc.{{.Name}}(value, in)
		if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, serialize.Response{Data: reply})

		return nil
	}
}
{{end}}