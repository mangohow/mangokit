package types

type Integer int

type UserInfoProto struct {
	Id         int            `json:"id" smapping:"id"`
	Username   string         `json:"username" smapping:"username"`
	Friends    []int          `json:"friends" smapping:"friends"`
	FriendsMap map[int]string `json:"friends_map" smapping:"friendsMap"`
	Age        Integer        `json:"age" smapping:"age"`
}

type UserInfo struct {
	Id       int    `json:"id" smapping:"id"`
	Username string `json:"username" smapping:"username"`
	Phone    string `json:"phone" smapping:"-"`
}
