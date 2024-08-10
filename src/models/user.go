package models

import "time"

type User struct {
	UserId      string
	UserName    string
	EmailId     string
	Posts       []Post
	Friends     []User
	ChatHistory [][]Message
}

type Message struct {
	MessageId string
	From      string
	To        string
	Content   string
	CreatedAt time.Time
}

type Post struct {
	PostId    string
	User      string
	Content   string
	CreatedAt time.Time
	// LastedUpdatedAt time.Time
	Tags []string
}
