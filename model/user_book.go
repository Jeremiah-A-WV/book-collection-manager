package model

import "time"

type UserBook struct {
	UserID int64
	BookID int64
	Status string
	Title  string
}

type UserBookDetail struct {
	BookID          int64     // from view.book_id
	Title           string    // view.title
	PublicationYear int       // view.publication_year
	ISBN            string    // view.isbn
	Status          string    // view.user_status
	Notes           string    // view.notes
	AddedAt         time.Time // view.added_at
	Authors         []string  // parsed from view.authors
	Genres          []string  // parsed from view.genres
	CreatedAt       time.Time // view.created_at
	UpdatedAt       time.Time // view.updated_at
}
