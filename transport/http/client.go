package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Client http client
type Client struct {
	client *http.Client
	config config
}

type config struct {
	host         string
	transport    http.RoundTripper
	interceptors []Interceptor
}

// Interceptor 拦截器
type Interceptor func() error

type ClientOption func(*config)

func NewClient(options ...ClientOption) (*Client, error) {
	c := &Client{
		client: &http.Client{},
	}
	for _, option := range options {
		option(&c.config)
	}

	c.client.Transport = c.config.transport

	if c.config.host == "" {
		return nil, errors.New("host is required, use WithHost option to set host")
	}

	return c, nil
}

func WithEndpoint(endpoint string) ClientOption {
	return func(c *config) {
		c.host = endpoint
	}
}

func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *config) {
		c.transport = transport
	}
}

func WithInterceptors(interceptors ...Interceptor) ClientOption {
	return func(c *config) {
		c.interceptors = append(c.interceptors, interceptors...)
	}
}

// Invoke 先执行全局拦截器，再执行CallOption中的before，最后再发起请求
func (c *Client) Invoke(ctx context.Context, method, path string, req, resp interface{}, opts ...CallOption) (status int, err error) {
	url := c.config.host + path

	bco := &BeforeCallInfo{
		Header: make(http.Header),
		Value:  req,
	}
	for _, opt := range opts {
		opt.Before(bco)
	}

	var bodyReader io.Reader
	if req != nil {
		bodyBytes, err := json.Marshal(req)
		if err != nil {
			return status, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}
	request, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return
	}

	if bco.ContentType == "" && method != http.MethodGet {
		request.Header.Set("Content-Type", "application/json")
	} else {
		request.Header.Set("Content-Type", bco.ContentType)
	}
	for k, v := range bco.Header {
		request.Header.Set(k, v[0])
	}

	response, err := c.client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	status = response.StatusCode

	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	if resp != nil && len(respBytes) > 0 && status >= 200 && status < 400 {
		if err = json.Unmarshal(respBytes, resp); err != nil {
			return
		}
	}

	aco := &AfterCallInfo{
		Resp:   resp,
		Status: status,
	}

	for _, opt := range opts {
		opt.After(aco)
	}

	return
}

func EncodeURL(pattern string, obj interface{}, query bool) string {
	strings.TrimSuffix(pattern, "/")
	if pattern == "" || obj == nil {
		return ""
	}

	var (
		builder = &strings.Builder{}
		encPam  = false
		m       = make(map[string]string, 8)
	)
	i := strings.IndexByte(pattern, ':')
	if i != -1 {
		// /xxx/:yyy/:zzz
		// /xxx
		builder.WriteString(pattern[:i-1])
		encPam = true
	} else {
		// /xxx/yyy
		// /xxx/yyy
		builder.WriteString(pattern)
	}
	// 从obj中获取param来拼接路径
	if encPam {
		// /:yyy/:zzz
		encodeParam(pattern[i-1:], obj, builder, m)
	}

	// 从obj中获取form参数来拼接路径
	// 如果对应的field字段为空，则不拼接
	if query {
		encodeQuery(obj, builder, m)
	}

	return builder.String()
}

func EncodeURLFromForm(pattern string, obj interface{}) string {
	if pattern == "" || obj == nil {
		return ""
	}
	builder := &strings.Builder{}
	builder.WriteString(pattern)
	encodeQuery(obj, builder, nil)

	return builder.String()
}

func encodeQuery(obj interface{}, builder *strings.Builder, expect map[string]string) {
	m := make(map[string]string)
	// 利用反射从结构体中获取form请求参数
	reflectGetValues(m, obj, FormKey)
	// 如果一个字段既添加了param tag 又添加了form tag，则需要忽略form tag，因为已经在param参数中进行了设置
	if expect != nil {
		for k := range m {
			if _, ok := expect[k]; ok {
				delete(m, k)
			}
		}
	}
	if len(m) == 0 {
		return
	}
	builder.WriteByte('?')
	for k, v := range m {
		if v != "" {
			builder.WriteString(k + "=" + v)
		}
	}
}

func encodeParam(pattern string, obj interface{}, builder *strings.Builder, m map[string]string) {
	paramPatterns := strings.Split(pattern, "/:")
	paramPatterns = removeEmptyStringsInPlace(paramPatterns)
	if len(paramPatterns) == 0 {
		return
	}

	for _, p := range paramPatterns {
		m[p] = ""
	}

	// 利用反射解析出param请求参数
	reflectGetValues(m, obj, ParamKey)
	for _, pp := range paramPatterns {
		if v := m[pp]; v != "" {
			builder.WriteByte('/')
			builder.WriteString(v)
		}
	}
}

func removeEmptyStringsInPlace(slice []string) []string {
	n := 0
	for _, str := range slice {
		if str != "" {
			slice[n] = str
			n++
		}
	}
	return slice[:n]
}

func reflectGetValues(m map[string]string, obj interface{}, tagK string) {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return
	}
	for rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	rt := rv.Type()
	n := rt.NumField()
	for i := 0; i < n; i++ {
		// 获取tag，如果没有则该tag不添加到路径中
		tagV := rt.Field(i).Tag.Get(tagK)
		if tagV == "" {
			continue
		}

		// 对于param参数，如果该tag没有在路径中声明，则不添加到路径中
		// 对于form参数，全部添加到路径中
		if _, ok := m[tagV]; !ok && tagK == ParamKey {
			continue
		}

		fiv := rv.Field(i)
		for fiv.Kind() == reflect.Ptr && !fiv.IsNil() {
			fiv = fiv.Elem()
		}
		if fiv.Kind() == reflect.Ptr && fiv.IsNil() {
			continue
		}

		switch fiv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			m[tagV] = strconv.FormatInt(fiv.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			m[tagV] = strconv.FormatUint(fiv.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			m[tagV] = strconv.FormatFloat(fiv.Float(), 'g', -1, 64)
		case reflect.Bool:
			m[tagV] = strconv.FormatBool(fiv.Bool())
		case reflect.String:
			m[tagV] = fiv.String()
		default:
		}
	}
}
