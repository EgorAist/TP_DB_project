package models

import (
	"time"
)

//easyjson:json
type RespError struct {
	Message string `json:"message"`
}

//easyjson:json
type Error struct {
	Code string `json:"-"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Code
}

//easyjson:json
type Forum struct {
	Slug string `json:"slug,omitempty"`
	Title string `json:"title,omitempty"`
	User string `json:"user,omitempty"`
	Threads int `json:"threads,omitempty"`
	Posts int `json:"posts,omitempty"`
}

//easyjson:json
type ForumCreate struct {
	Slug string `json:"slug"`
	Title string `json:"title"`
	User string `json:"user"`
}

type ForumInput struct {
	Slug string
}

type ForumGetUsers struct {
	Slug string
	Limit int
	Since string
	Desc bool
}

type ForumGetThreads struct {
	Slug string
	Limit int
	Since string
	Desc bool
}

type UserInput struct {
	Nickname string
}

//easyjson:json
type User struct {
	Nickname string `json:"nickname,omitempty"`
	Fullname string `json:"fullname,omitempty"`
	Email string `json:"email,omitempty"`
	About string `json:"about,omitempty"`
}

//easyjson:json
type Thread struct {
	Author  string    `json:"author,omitempty"`
	Created time.Time `json:"created,omitempty"`
	Forum   string    `json:"forum,omitempty"`
	ID      int       `json:"id,omitempty"`
	Message string    `json:"message,omitempty"`
	Slug    string    `json:"slug,omitempty"`
	Title   string    `json:"title,omitempty"`
	Votes   int       `json:"votes,omitempty"`
}

//easyjson:json
type ThreadInput struct {
	ThreadID int    `json:"thread"`
	Slug     string `json:"-"`
}

//easyjson:json
type ThreadUpdate struct {
	ThreadInput
	Title    string `json:"title"`
	Message  string `json:"message"`
}

type ThreadGetPosts struct {
	ThreadInput
	Limit int
	Since int
	Sort string
	Desc bool
}

type PostInput struct {
	ID       int  `json:"id"`
}

//easyjson:json
type PostUpdate struct {
	ID       int  `json:"id"`
	Message string `json:"message"`
}
//easyjson:json
type PostCreate struct {
	Parent   int  `json:"parent,omitempty"`
	Author   string `json:"author,omitempty"`
	Message  string `json:"message,omitempty"`
}
//easyjson:json
type Post struct {
	ThreadInput
	ID       int  `json:"id,omitempty"`       // Идентификатор данного сообщения.
	Parent   int  `json:"parent,omitempty"`   // Идентификатор родительского сообщения (0 - корневое сообщение обсуждения).
	Author   string `json:"author,omitempty"`   // Автор, написавший данное сообщение.
	Message  string `json:"message,omitempty"`  // Собственно сообщение форума.
	IsEdited bool   `json:"isEdited,omitempty"` // Истина, если данное сообщение было изменено.
	Forum    string `json:"forum,omitempty"`    // Идентификатор форума (slug) данного сообещния.
	Created  string `json:"created,omitempty"`
}

//easyjson:json
type PostFull struct {
	Author *User `json:"author,omitempty"`
	Forum *Forum `json:"forum,omitempty"`
	Post *Post `json:"post,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
}

//easyjson:json
type Vote struct {
	User string `json:"nickname"`
	Voice int `json:"voice"`
	Thread ThreadInput `json:"_"`
}

//easyjson:json
type Status struct {
	Forum  int32 `json:"forum"`
	Post   int64 `json:"post"`
	Thread int32 `json:"thread"`
	User   int32 `json:"user"`
}