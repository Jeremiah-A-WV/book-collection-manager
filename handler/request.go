package handler

import (
	"book-collection-manager/config"
	"book-collection-manager/middleware"
	"book-collection-manager/model"
	"book-collection-manager/repository"
	"html/template"
	"net/http"
)

func BookRequestFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/request.html"))
	tmpl.Execute(w, nil)
}

func BookRequestHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetSessionUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	req := &model.BookRequest{
		Title:      r.FormValue("title"),
		AuthorName: r.FormValue("author_name"),
		ISBN:       r.FormValue("isbn"),
		UserID:     userID,
	}

	err := repository.CreateBookRequest(config.DB, req)
	if err != nil {
		http.Error(w, "Error submitting request", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
