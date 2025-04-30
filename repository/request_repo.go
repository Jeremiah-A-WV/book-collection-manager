package repository

import (
	"book-collection-manager/model"
	"database/sql"
)

func CreateBookRequest(db *sql.DB, req *model.BookRequest) error {
	_, err := db.Exec(`INSERT INTO book_request (user_id, title, author_name, isbn) VALUES (?, ?, ?, ?)`,
		req.UserID, req.Title, req.AuthorName, req.ISBN)
	return err
}
