package handler

import (
	"book-collection-manager/config"
	"book-collection-manager/repository"
	"html/template"
	"net/http"
	"strings"
)

func BookListHandler(w http.ResponseWriter, r *http.Request) {
	books, err := repository.GetAllBooks(r.Context(), config.DB)
	if err != nil {
		http.Error(w, "Failed to load books", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(
		template.New("books.html").
			Funcs(template.FuncMap{"join": strings.Join}).
			ParseFiles("templates/books.html"),
	)
	// passing the slice directly, so "{{ range . }}" is correct
	tmpl.Execute(w, books)
}
