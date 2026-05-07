package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Nobodywinsbutme/mangahub/internal/database"
	"github.com/Nobodywinsbutme/mangahub/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTSecret should come from config in production.
// For now, hardcode it — we'll move it to config.yaml in Phase 3.
const JWTSecret = "your-secret-key-change-this"

// generateID creates a simple unique ID.
// In production you'd use a UUID library, but fmt.Sprintf works for now.
func generateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func RegisterUser(username, email, password string) (*models.User, error) {
	// 1. Hash the password — bcrypt cost 12 is a good balance of security/speed
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           generateID("usr"),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
	}

	// 2. Insert into DB
	_, err = database.DB.Exec(
		`INSERT INTO users (id, username, email, password_hash) VALUES (?, ?, ?, ?)`,
		user.ID, user.Username, user.Email, user.PasswordHash,
	)
	if err != nil {
		// SQLite returns an error containing "UNIQUE constraint failed" on duplicates
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	return user, nil
}

func LoginUser(username, password string) (string, *models.User, error) {
	user := &models.User{}

	// 1. Fetch user from DB
	row := database.DB.QueryRow(
		`SELECT id, username, email, password_hash FROM users WHERE username = ?`,
		username,
	)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if err == sql.ErrNoRows {
		return "", nil, errors.New("account not found")
	}
	if err != nil {
		return "", nil, err
	}

	// 2. Compare password with stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// 3. Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}
