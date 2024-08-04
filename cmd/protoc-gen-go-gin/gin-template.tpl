
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
    {{- end -}}
}
{{end}}

type {{.ServiceName}}HTTPClient interface {
{{- range .Methods}}
	{{- if and (eq .InputFieldLen 0) (eq .OutputFieldLen 0)}}
        {{.Name}}(ctx context.Context, opts ...http.CallOption) error
    {{- else if eq .InputFieldLen 0}}
        {{.Name}}(ctx context.Context, opts ...http.CallOption) (*{{.Reply}}, error)
    {{- else if eq .OutputFieldLen 0}}
        {{.Name}}(ctx context.Context, req *{{.Request}}, opts ...http.CallOption) error
    {{- else}}
        {{.Name}}(ctx context.Context, req *{{.Request}}, opts ...http.CallOption) (*{{.Reply}}, error)
    {{- end}}
{{- end}}
}

type {{.LowerServiceName}}HTTPClient struct {
	cc *http.Client
}

func New{{.ServiceName}}HTTPClient(client *http.Client) {{.ServiceName}}HTTPClient {
	return &{{.LowerServiceName}}HTTPClient{cc: client}
}

{{range .Methods}}
{{- if and (ne .InputFieldLen 0) (ne .OutputFieldLen 0) -}}
func (c *{{.LowerServiceName}}HTTPClient) {{.Name}}(ctx context.Context, req *{{.Request}}, opts ...http.CallOption) (*{{.Reply}}, error) {
{{- else if ne .InputFieldLen 0 -}}
func (c *{{.LowerServiceName}}HTTPClient) {{.Name}}(ctx context.Context, req *{{.Request}}, opts ...http.CallOption) error {
{{- else if ne .OutputFieldLen 0 -}}
func (c *{{.LowerServiceName}}HTTPClient) {{.Name}}(ctx context.Context, opts ...http.CallOption) (*{{.Reply}}, error) {
{{- else -}}
func (c *{{.LowerServiceName}}HTTPClient) {{.Name}}(ctx context.Context, opts ...http.CallOption) error {
{{- end -}}
    {{- if ne .OutputFieldLen 0}}
	reply := new({{.Reply}})
    {{- end}}
    {{- if and .EncodeParam .EncodeForm}}
	pattern := "{{.Path}}"
    path := http.EncodeURL(pattern, req, true)
    {{- else if .EncodeParam}}
    pattern := "{{.Path}}"
    path := http.EncodeURL(pattern, req, false)
    {{- else if .EncodeForm}}
    pattern := "{{.Path}}"
    path := http.EncodeURLFromForm(pattern, req)
    {{- else}}
    path := "{{.Path}}"
    {{- end}}
	{{- if and (ne .InputFieldLen 0) (ne .OutputFieldLen 0)}}
    _, err := c.cc.Invoke(ctx, "{{.Method}}", path, req, reply, opts...)
    {{- else if ne .InputFieldLen 0}}
    _, err := c.cc.Invoke(ctx, "{{.Method}}", path, req, nil, opts...)
    {{- else if ne .OutputFieldLen 0}}
    _, err := c.cc.Invoke(ctx, "{{.Method}}", path, nil, reply, opts...)
    {{- else}}
    _, err := c.cc.Invoke(ctx, "{{.Method}}", path, nil, nil, opts...)
    {{- end}}
	
    {{if ne .OutputFieldLen 0}}
	return reply, err
    {{else}}
    return err
    {{- end -}}
}
{{end}}

var _{{.ServiceName}}HTTPService_serviceDesc = &http.ServiceDesc{
	HandlerType: (*{{.ServiceName}}HTTPService)(nil),
	Methods: []http.MethodDesc{
	{{- range .Methods}}
		{
			Method:  "{{.Method}}",
			Path:    "{{.Path}}",
			Handler: _{{.ServiceName}}_{{.Name}}_HTTP_Handler,
		},
	{{- end}}
	},
}
