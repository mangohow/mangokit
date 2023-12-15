package httpwrapper

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var ctxPool = sync.Pool{
	New: func() any {
		return &Context{}
	},
}

func putPool(ctx *Context) {
	ctxPool.Put(ctx)
}

type Context struct {
	*gin.Context
	fnt DecodeFieldNameType
}

func (c *Context) clear() {
	c.fnt = 0
	c.Context = nil
}

func (c *Context) bindUri(val interface{}) (err error) {
	if len(c.Params) == 0 {
		return nil
	}

	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}

	return c.mapRequest(val, m)
}

func (c *Context) bindQuery(val interface{}) (err error) {
	values := c.Request.URL.Query()
	if values == nil {
		return nil
	}

	return c.mapRequest(val, values)
}

func (c *Context) mapRequest(val interface{}, form map[string][]string) error {
	// 从json tag中解析
	if c.fnt == JsonTag {
		return binding.MapFormWithTag(val, form, "json")
	}

	// 根据名称解析, 修改form中的key，指定空tag，则会使用字段名称进行解析
	kvs := make(map[string][]string)
	for k, v := range form {
		name := c.fnt.ToFieldName(k)
		kvs[name] = v
	}

	return binding.MapFormWithTag(val, kvs, "")
}

func (c *Context) BindRequest(val interface{}) (err error) {
	// bind body
	if err = c.ShouldBind(val); err != nil {
		return err
	}

	// bind query
	if err = c.bindQuery(val); err != nil {
		return err
	}

	// bind uri
	err = c.bindUri(val)

	return
}
