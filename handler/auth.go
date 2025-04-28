package handler

import (
	"book-collection-manager/config"
	"book-collection-manager/middleware"
	"book-collection-manager/model"
	"book-collection-manager/repository"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func GenerateRandomSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func HashPassword(password string) (string, error) {
	salt, err := GenerateRandomSalt(16) // 16 bytes salt
	if err != nil {
		return "", err
	}

	// Argon2id Parameters
	time := uint32(1)           // number of iterations
	memory := uint32(64 * 1024) // 64 MB memory
	threads := uint8(4)         // number of threads
	keyLen := uint32(32)        // desired hash length

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Final encoded string:
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memory, time, threads, b64Salt, b64Hash)

	return encoded, nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var memory uint32
	var time uint32
	var threads uint8

	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	// Compute hash of incoming password
	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(hash)))

	// Compare byte slices
	if subtle.ConstantTimeCompare(hash, computedHash) == 1 {
		return true, nil
	}
	return false, nil
}

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
		tmpl := template.Must(template.ParseFiles("templates/books.html"))
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

	// Fetch the list of books
	books, err := repository.GetAllBooks(r.Context(), config.DB)
	if err != nil {
		http.Error(w, "Failed to load books", http.StatusInternalServerError)
		return
	}

	// Render books.html directly
	tmpl := template.Must(template.ParseFiles("templates/books.html"))
	tmpl.Execute(w, books)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login.html"))

	// Handle login form submission
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := repository.AuthenticateUser(r.Context(), email, password)
		if err != nil {
			// Authentication failed – render login page with error
			tmpl.Execute(w, map[string]interface{}{
				"Error": "Invalid email or password.",
			})
			return
		}

		// Set session and render the book list page instead of redirect
		middleware.SetSession(w, user.ID)
		BookListHandler(w, r.WithContext(r.Context()))
		return
	}

	// Default: render login page
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
