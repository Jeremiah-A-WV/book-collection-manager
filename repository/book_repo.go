package repository

import (
	"book-collection-manager/model"
	"context"
	"database/sql"
	"strings"
	"time"
)

func GetAllBooks(ctx context.Context, db *sql.DB) ([]model.Book, error) {
	const sqlQuery = `
        SELECT 
            id,
            title,
            publication_year,
            isbn,
            authors,
            genres,
            created_at,
            updated_at
        FROM vw_books_detailed
        WHERE 1 = 1
    `

	rows, err := db.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []model.Book
	for rows.Next() {
		var b model.Book
		var authorsCSV, genresCSV string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(
			&b.ID,
			&b.Title,
			&b.PublicationYear,
			&b.ISBN,
			&authorsCSV,
			&genresCSV,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		// Split the comma-separated lists into slices
		if authorsCSV == "" {
			b.Authors = []string{}
		} else {
			b.Authors = strings.Split(authorsCSV, ",")
		}
		if genresCSV == "" {
			b.Genres = []string{}
		} else {
			b.Genres = strings.Split(genresCSV, ",")
		}

		b.CreatedAt = createdAt
		b.UpdatedAt = updatedAt

		books = append(books, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return books, nil
}
