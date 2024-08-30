package models

import "time"

type User struct {
	UserId        string
	UserName      string
	EmailId       string
	Visibility    string
	Posts         []Post
	Friends       []User
	Notifications []Notification
	ChatHistory   [][]Message
}

type Notification struct {
	NID       string
	IsRead    bool
	TimeStamp int64
	Content   string
	CType     string // one of ("connReq", "connAccepted", "media")
	MetaData  map[string]string
}

type ConnectionRequest struct {
	From       string
	To         string
	ConnStatus string
}

type Message struct {
	MessageId string
	From      string
	To        string
	Content   string
	CreatedAt time.Time
}

type Post struct {
	PostId          string
	User            string
	Content         string
	CreatedAt       time.Time
	LastedUpdatedAt time.Time
	Tags            []string
}
