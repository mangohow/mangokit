package httpwrapper

import "strings"

// DecodeFieldNameType 由于proto文件生成的结构体中的tag只有protobuf和json
// 因此在解析query或uri类型的参数时需要指定反射时的字段名称对应的类型
type DecodeFieldNameType int8

const (
	// JsonTag 根据json tag解析
	JsonTag DecodeFieldNameType = iota

	// CamelCase 小驼峰 CreateTime --> createTime
	CamelCase

	// PascalCase 大驼峰 CreateTime --> CreateTime
	PascalCase

	// SnakeCase 下划线 CreateTime --> create_time
	SnakeCase
)

// ToFieldName 将请求字段名称转换为大驼峰的字段名称
func (d DecodeFieldNameType) ToFieldName(name string) string {
	if len(name) == 0 {
		return name
	}

	switch d {
	case CamelCase:
		return title(name)
	case SnakeCase:
		words := strings.Split(name, "_")
		for i := 0; i < len(words); i++ {
			words[i] = title(words[i])
		}
		return strings.Join(words, "")
	}

	return name
}

func title(str string) string {
	if len(str) == 0 {
		return ""
	}

	builder := strings.Builder{}
	builder.Grow(len(str))
	builder.WriteByte(str[0] & '_')
	builder.WriteString(str[1:])
	return builder.String()
}
