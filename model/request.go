package model

type BookRequest struct {
	UserID     int64
	Title      string
	AuthorName string
	ISBN       string
}
