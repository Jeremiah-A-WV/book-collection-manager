package main

import (
	"book-collection-manager/config"
	"book-collection-manager/handler"
	"book-collection-manager/middleware"
	"log"
	"net/http"
)

func main() {
	// Initialize DB
	config.LoadDBConfig() // Load the database configuration and initialize the connection

	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Auth
	mux.HandleFunc("/login", handler.LoginHandler)
	mux.HandleFunc("/logout", handler.LogoutHandler)
	mux.HandleFunc("/guest", handler.GuestLoginHandler)
	mux.HandleFunc("/register", handler.RegisterHandler)

	// Book list
	mux.HandleFunc("/books", middleware.WithSession(handler.BookListHandler))
	mux.HandleFunc("/add-to-list", middleware.WithSession(handler.AddToUserListHandler))

	// Book requests
	mux.HandleFunc("/request", middleware.WithSession(handler.BookRequestFormHandler))
	mux.HandleFunc("/submit-request", middleware.WithSession(handler.BookRequestHandler))

	// Personal book list
	mux.HandleFunc("/myList", middleware.WithSession(handler.UserBookListHandler))
	mux.HandleFunc("/remove-from-list", middleware.WithSession(handler.RemoveFromUserListHandler))

	// Profile
	mux.HandleFunc("/profile", middleware.WithSession(handler.UpdateProfileHandler))

	log.Println("Server running on http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
