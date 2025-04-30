package repository

import (
	"book-collection-manager/auth"
	"book-collection-manager/config"
	"book-collection-manager/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
)

func CreateNewUser(ctx context.Context, user *model.User) (err error) {
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Hash the password before saving
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO user (username, email, password, is_active, is_guest)
		VALUES (?, ?, ?, TRUE, ?)
	`
	result, err := tx.ExecContext(ctx, query, user.Username, user.Email, hashedPassword, user.IsGuest)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	user.Password = "" // clear plaintext password
	return nil
}

func CreateGuestUser(ctx context.Context) (*model.User, error) {
	user := &model.User{
		Username: "guest_" + uuid.New().String(),
		Email:    fmt.Sprintf("guest_%s@example.com", uuid.NewString()),
		IsGuest:  true,
	}
	err := CreateNewUser(ctx, user)
	log.Printf("Creating user: username=%s email=%s", user.Username, user.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByUsername(db *sql.DB, username string) (*model.User, error) {
	var user model.User
	err := db.QueryRow("SELECT id, username, email, password, is_guest FROM user WHERE username = ?", username).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsGuest)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(db *sql.DB, id int64) (*model.User, error) {
	var user model.User
	err := db.QueryRow("SELECT id, username, email, password, is_guest FROM user WHERE id = ?", id).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsGuest)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func DeleteUserByID(ctx context.Context, id int64) error {
	query := `DELETE FROM user WHERE id = ?`

	_, err := config.DB.ExecContext(ctx, query, id)
	log.Printf("Deleting user: id=%s", id)
	if err != nil {
		return err
	}

	return nil
}

func UpdateUserProfile(ctx context.Context, user *model.User) error {
	query := `
		UPDATE user
		SET username = ?, email = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := config.DB.ExecContext(ctx, query, user.Username, user.Email, user.ID)
	return err
}

func UpdateUserPassword(ctx context.Context, userID int64, newPassword string) error {
	query := `
		UPDATE user
		SET password = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := config.DB.ExecContext(ctx, query, newPassword, userID)
	return err
}

func AuthenticateUser(ctx context.Context, email string, password string) (*model.User, error) {
	var user model.User
	query := `SELECT id, username, email, password, is_active, is_guest FROM user WHERE email = ?`
	err := config.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.IsActive,
		&user.IsGuest,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Verify password
	ok, err := auth.VerifyPassword(password, user.Password)
	if err != nil || !ok {
		return nil, errors.New("invalid email or password")
	}

	// Stops from revealing password
	user.Password = ""

	return &user, nil
}
