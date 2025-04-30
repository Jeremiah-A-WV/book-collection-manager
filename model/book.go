package model

import "time"

type Book struct {
	ID              int64
	Title           string
	Description     string
	PublicationYear int
	ISBN            string
	Authors         []string
	Genres          []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
