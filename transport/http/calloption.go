package http

import "net/http"

type BeforeCallInfo struct {
	ContentType string
	Header      http.Header
	Value       interface{}
}

type AfterCallInfo struct {
	Resp   interface{}
	Status int
}

// CallOption 在请求被调用前和调用后执行的handler
type CallOption interface {
	Before(call *BeforeCallInfo)

	After(call *AfterCallInfo)
}

type EmptyCallOptions struct{}

func (EmptyCallOptions) Before(*BeforeCallInfo) {}
func (EmptyCallOptions) After(*AfterCallInfo)   {}

type contentTypeCallOption struct {
	EmptyCallOptions
	contentType string
}

func (c contentTypeCallOption) Before(info *BeforeCallInfo) {
	info.ContentType = c.contentType
}

// ContentTypeCallOption 为请求设置content type
func ContentTypeCallOption(contentType string) CallOption {
	return contentTypeCallOption{contentType: contentType}
}

type headerCallOption struct {
	EmptyCallOptions
	headers http.Header
}

func (c headerCallOption) Before(info *BeforeCallInfo) {
	info.Header = c.headers
}

// HeadersCallOption 为请求设置headers
func HeadersCallOption(headers http.Header) CallOption {
	return headerCallOption{headers: headers}
}
