package model

type UserProto struct {
	Id int
}

type Id struct {
	Id int
}

type Username struct {
	Username string
}

type User struct {
	Id       int
	Username string
	U        Username
}

type UserInfo struct {
	U     User
	Email string
}

type Integer int
