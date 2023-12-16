package httpwrapper

import (
	"fmt"
	"reflect"
	"testing"
)

func TestToFieldName(t *testing.T) {
	cc := CamelCase
	fmt.Println(cc.ToFieldName("c"))
	fmt.Println(cc.ToFieldName("createTime"))

	pc := PascalCase
	fmt.Println(pc.ToFieldName("C"))
	fmt.Println(pc.ToFieldName("CreateTime"))

	sc := SnakeCase
	fmt.Println(sc.ToFieldName("c"))
	fmt.Println(sc.ToFieldName("create_time"))
}

type User struct {
	Username string `json:"username"`
}

func TestGetTag(t *testing.T) {
	u := User{Username: "abc"}
	typ := reflect.TypeOf(u)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		t.Log("tag:", field.Tag, "json tag:", field.Tag.Get("json"), "empty tag:", field.Tag.Get(""))

	}
}
