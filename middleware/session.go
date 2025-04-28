package middleware

import (
	"book-collection-manager/config"
	"book-collection-manager/repository"
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var sessions = make(map[string]int64) // sessionID -> userID

const SessionCookieName = "session_id"

// SetSession sets a session cookie with the user ID
func SetSession(w http.ResponseWriter, userID int64) {
	sessionID := uuid.NewString()
	sessions[sessionID] = userID

	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, cookie)
}

// GetSessionUserID retrieves the user ID from the session
func GetSessionUserID(r *http.Request) (int64, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return 0, false
	}
	userID, exists := sessions[cookie.Value]
	return userID, exists
}

// ClearSession clears the session for the current user
func ClearSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		delete(sessions, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(-1 * time.Hour),
	})
}

// WithSession ensures the user is authenticated for each request
func WithSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve user ID from session
		userID, ok := GetSessionUserID(r)
		if !ok {
			// If no session found, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Retrieve user data from the database
		user, err := repository.GetUserByID(config.DB, userID)
		if err != nil {
			// If user not found, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Attach user data to request
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// SessionMiddleware is a general session handler for non-authenticated routes
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Continue with the request without any additional session check for now
		next.ServeHTTP(w, r)
	})
}
