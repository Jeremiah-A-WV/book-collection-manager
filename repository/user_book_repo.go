package repository

import (
	"book-collection-manager/config"
	"book-collection-manager/model"
	"context"
	"database/sql"
	"errors"
	"strings"
)

func AddBookToUserList(ctx context.Context, userID, bookID int64) error {
	_, err := config.DB.ExecContext(ctx,
		"INSERT INTO user_book (user_id, book_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE book_id = book_id",
		userID, bookID,
	)
	return err
}

func GetUserBooks(ctx context.Context, userID int64) ([]model.UserBookDetail, error) {
	const sqlQuery = `
    SELECT
      book_id,
      title,
      publication_year,
      isbn,
      user_status,
      notes,
      added_at,
      authors,
      genres,
      created_at,
      updated_at
    FROM vw_user_books_detailed
    WHERE user_id = ?
    `

	rows, err := config.DB.QueryContext(ctx, sqlQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.UserBookDetail
	for rows.Next() {
		var ub model.UserBookDetail

		// scan nullable columns into sql.NullString
		var (
			isbnNull    sql.NullString
			notesNull   sql.NullString
			authorsNull sql.NullString
			genresNull  sql.NullString
		)

		if err := rows.Scan(
			&ub.BookID,
			&ub.Title,
			&ub.PublicationYear,
			&isbnNull,
			&ub.Status,
			&notesNull,
			&ub.AddedAt,
			&authorsNull,
			&genresNull,
			&ub.CreatedAt,
			&ub.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// convert NullString -> string
		if isbnNull.Valid {
			ub.ISBN = isbnNull.String
		}
		if notesNull.Valid {
			ub.Notes = notesNull.String
		}
		// split comma-lists
		if authorsNull.Valid && authorsNull.String != "" {
			ub.Authors = strings.Split(authorsNull.String, ",")
		}
		if genresNull.Valid && genresNull.String != "" {
			ub.Genres = strings.Split(genresNull.String, ",")
		}

		list = append(list, ub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func RemoveBookFromUserList(ctx context.Context, userID, bookID int64) error {
	result, err := config.DB.ExecContext(ctx,
		"DELETE FROM user_book WHERE user_id = ? AND book_id = ?",
		userID, bookID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("no rows deleted")
	}
	return nil
}
