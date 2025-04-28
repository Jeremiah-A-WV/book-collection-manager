package handler

import (
	"book-collection-manager/config"
	"book-collection-manager/middleware"
	"book-collection-manager/model"
	"book-collection-manager/repository"
	"html/template"
	"net/http"
)

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetSessionUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		newUsername := r.FormValue("username")
		newEmail := r.FormValue("email")
		newPassword := r.FormValue("password")

		user := &model.User{
			ID:       userID,
			Username: newUsername,
			Email:    newEmail,
		}

		err := repository.UpdateUserProfile(r.Context(), user)
		if err != nil {
			http.Error(w, "Error updating profile", http.StatusInternalServerError)
			return
		}

		if newPassword != "" {
			err = repository.UpdateUserPassword(r.Context(), userID, newPassword)
			if err != nil {
				http.Error(w, "Error updating password", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	user, err := repository.GetUserByID(config.DB, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/profile.html"))
	tmpl.Execute(w, user)
}
