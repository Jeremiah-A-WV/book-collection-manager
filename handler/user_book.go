package handler

import (
	"book-collection-manager/middleware"
	"book-collection-manager/repository"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func AddToUserListHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetSessionUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	temp := r.FormValue("book_id")
	bookID, err1 := strconv.ParseInt(temp, 10, 64)
	if err1 != nil {
		fmt.Println("Error:", err1)
		return
	}

	err := repository.AddBookToUserList(r.Context(), userID, bookID)
	if err != nil {
		http.Error(w, "Could not add book to list", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/myList", http.StatusSeeOther)
}

func UserBookListHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetSessionUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	list, err := repository.GetUserBooks(r.Context(), userID)
	if err != nil {
		log.Printf("ERROR in GetUserBooks: %v", err)
		http.Error(w, "Error fetching your books", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(
		template.New("myList.html").
			Funcs(template.FuncMap{
				"statusClass": func(status string) string {
					switch status {
					case "TO_READ":
						return "to-read"
					case "READING":
						return "reading"
					case "READ":
						return "read"
					case "ABANDONED":
						return "abandoned"
					default:
						return ""
					}
				},
			}).
			ParseFiles("templates/myList.html"),
	)

	if err := tmpl.Execute(w, list); err != nil {
		log.Printf("ERROR executing template: %v", err)
		return
	}
}

func RemoveFromUserListHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetSessionUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	bookID, err := strconv.ParseInt(r.FormValue("book_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	if err := repository.RemoveBookFromUserList(r.Context(), userID, bookID); err != nil {
		http.Error(w, "Failed to remove book", http.StatusInternalServerError)
		return
	}

	// Send empty content to replace the book's <li>
	w.WriteHeader(http.StatusNoContent)
}
