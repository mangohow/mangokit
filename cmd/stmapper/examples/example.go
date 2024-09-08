package examples

import (
	"github.com/mangohow/mangokit/cmd/stmapper/examples/model"
	"github.com/mangohow/mangokit/stmapper"
	"time"
)

type TestType string

type TypeTestType TestType

type MyInteger model.Integer

type MyIntegerPointer *model.Integer

type OuterStruct model.Username

type OuterStructPointer *model.Username

type St struct {
	XXX string
}

type Interface interface {
	AAA()
}

type UserProto struct {
	// 基本类型
	//	Id int `stmapper:"id"`
	// 基本类型指针
	//	IdP *int
	// 底层是基本类型的自定义类型
	//	A TestType `stmapper:"a"`
	// 底层是基本类型的自定义类型指针
	//	AA *TestType `stmapper:"aa"`
	// 切片类型
	//	AAA []TestType `stmapper:"aaa"`
	// 元素为指针的切片
	//	AAAA []*TestType `stmapper:"aaaa"`
	// 元素为空接口的切片
	//	AAAAA []interface{}
	// 元素为空接口指针的切片
	//	AAAAAA []*interface{}
	// 元素为空接口别名的切片
	//	AAAAAAA []any
	// 元素为空接口别名指针的切片
	//	AAAAAAAA []*any
	// 结构体类型
	//	B St
	// 结构体类型指针
	//	C *St
	// 元素为基本类型的切片
	//	D []string
	// 元素为基本类型指针的切片
	//	E []*string
	//	EE  *[]string
	//	EEE *[]*string
	// 外部结构体类型
	//	F model.Username
	// 外部结构体指针类型
	//	G *model.Username
	// 元素为外部结构体的切片
	//	H []model.Username
	// 元素为外部结构体指针的切片
	//	I []*model.Username
	// 空接口
	//	J interface{}
	// 空接口指针
	//	K *interface{}
	// 空接口别名
	//	L any
	// 空接口别名指针
	//	M *any
	// 外部接口
	//	N fmt.Stringer
	// 外部接口指针
	//	O *fmt.Stringer
	// 接口
	//	P Interface
	// 接口指针
	//	Q *Interface
	// 从外部类型定义的基本类型
	//	R MyInteger
	// 从外部类型定义的基本类型指针
	//	S *MyInteger
	// 从外部类型定义的指针基本类型
	//	T MyIntegerPointer
	// 从外部类型定义的指针基本类型指针
	//	U *MyIntegerPointer
	// 从外部结构体定义的类型
	//	V OuterStruct
	// 从外部结构体定义的类型指针
	//	W *OuterStruct
	// 从外部结构体定义的指针类型
	//	X OuterStructPointer
	// 从外部结构体定义的指针类型指针
	//	Y *OuterStructPointer
	// 时间类型
	TT  time.Time
	TTT *time.Time
}

type User struct {
	Id int `stmapper:"id"`
}

// Conv3 将结构体id和username映射到类型为User的结构体，并返回
func Conv3(id model.Id, username model.Username) (u model.User) {
	panic(stmapper.BuildMappingFrom(id, username))
}

// Conv4 将结构体id和username映射到类型为User的结构体
func Conv4(id model.Id, username model.Username) (u *model.User) {
	panic(stmapper.BuildMappingFrom(id, username))
}

func Conv(up *UserProto, u *User) {
	panic(stmapper.BuildMapping(up, u))
}

// Conv1 将up结构体字段映射到u
func Conv1(up model.UserProto, u *model.User) {
	panic(stmapper.BuildMapping(up, u))
}

// Conv2 将up结构体映射到类型为User的结构体，并返回
func Conv2(up model.UserProto) model.User {
	panic(stmapper.BuildMappingFrom(up))
}

// UserProtoToUser1 赋值是无用的
func UserProtoToUser1(u1 model.UserProto, u2 model.User) {
	u2.Id = u1.Id
}

// UserProtoToUser2 赋值是无用的
func UserProtoToUser2(u1 *model.UserProto, u2 model.User) {
	u2.Id = u1.Id
}

func UserProtoToUser3(u1 model.UserProto, u2 *model.User) {
	u2.Id = u1.Id
}

func UserProtoToUser4(u1 *model.UserProto, u2 *model.User) {
	u2.Id = u1.Id
}

func ToUser1(u model.User) model.User {
	return model.User{
		Id: u.Id,
	}
}

func ToUser2(u *model.User) model.User {
	return model.User{
		Id: u.Id,
	}
}

func ToUser3(u model.User) *model.User {
	return &model.User{
		Id: u.Id,
	}
}

func ToUser4(u *model.User) *model.User {
	return &model.User{
		Id: u.Id,
	}
}

func ToUserSlice1(us []model.UserProto) []model.User {
	res := make([]model.User, len(us))
	for i := range us {
		res[i] = model.User{
			Id: us[i].Id,
		}
	}

	return res
}

func ToUserSlice2(us []*model.UserProto) []model.User {
	res := make([]model.User, len(us))
	for i := range us {
		res[i] = model.User{
			Id: us[i].Id,
		}
	}

	return res
}

func ToUserSlice3(us []model.UserProto) []*model.User {
	res := make([]*model.User, len(us))
	for i := range us {
		res[i] = &model.User{
			Id: us[i].Id,
		}
	}

	return res
}

func ToUserSlice4(us []*model.UserProto) []*model.User {
	res := make([]*model.User, len(us))
	for i := range us {
		res[i] = &model.User{
			Id: us[i].Id,
		}
	}

	return res
}

func ToUserSlice5(us []*model.UserProto) (u []*model.User) {
	u = make([]*model.User, len(us))
	for i := range us {
		u[i] = &model.User{
			Id: us[i].Id,
		}
	}

	return u
}
