package handler

import (
	"book-collection-manager/config"
	"book-collection-manager/middleware"
	"book-collection-manager/model"
	"book-collection-manager/repository"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()

		user := &model.User{
			Username: r.FormValue("username"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}

		err := repository.CreateNewUser(r.Context(), user)
		if err != nil {
			http.Error(w, "Could not create user", http.StatusInternalServerError)
			return
		}

		// Log the user in
		middleware.SetSession(w, user.ID)

		// Load user books
		books, err := repository.GetAllBooks(r.Context(), config.DB)
		if err != nil {
			http.Error(w, "Could not load books", http.StatusInternalServerError)
			return
		}

		// Render books.html directly
		tmpl := template.Must(
			template.New("books.html").
				Funcs(template.FuncMap{"join": strings.Join}).
				ParseFiles("templates/books.html"),
		)
		tmpl.Execute(w, books)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/register.html"))
	tmpl.Execute(w, nil)
}

func GuestLoginHandler(w http.ResponseWriter, r *http.Request) {
	user, err := repository.CreateGuestUser(r.Context())
	if err != nil {
		log.Printf("Failed to create guest user: %v", err)
		http.Error(w, "Failed to create guest user", http.StatusInternalServerError)
		return
	}
	log.Printf("Guest login triggered: userID=%d", user.ID)
	middleware.SetSession(w, user.ID)

	// Redirect guests to the main book list page
	http.Redirect(w, r, "/books", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login.html"))

	// Handle login form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := repository.AuthenticateUser(r.Context(), email, password)
		if err != nil {
			tmpl.Execute(w, map[string]interface{}{
				"Error": "Invalid email or password.",
			})
			return
		}

		// Set the session cookie
		middleware.SetSession(w, user.ID)

		// Redirect to the "All Books page
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	// Render login page
	tmpl.Execute(w, nil)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	userID, ok := middleware.GetSessionUserID(r)
	if !ok {
		// If no user is logged in, render the login page (or any other page)
		http.ServeFile(w, r, "templates/login.html")
		return
	}

	// Check if the user is a guest
	user, err := repository.GetUserByID(config.DB, userID)
	if err != nil {
		http.Error(w, "Error retrieving user", http.StatusInternalServerError)
		return
	}

	// If the user is a guest, delete them
	if user.IsGuest {
		err := repository.DeleteUserByID(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "Failed to delete guest user", http.StatusInternalServerError)
			return
		}
	}

	// Clear the session
	middleware.ClearSession(w, r)

	// Render a logout confirmation page
	tmpl := template.Must(template.ParseFiles("templates/logout.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Message": "You have successfully logged out.",
	})
}
