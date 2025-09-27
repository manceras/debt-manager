package handlers

import (
	"debt-manager/internal/db"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
)

type CreateUserRequest struct {
	Username string
	Email    string
	Password string
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func containsRestrictedChars(s string) bool {
	var restricted_chars = []string{" ", "/", "\\", "?", "%", "*", ":", "|", "\"", "<", ">"}
	for _, char := range restricted_chars {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}

func formatDBUser(user db.AppUsersSafe) map[string]interface{} {
	return map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"created_at": user.CreatedAt,
		"last_login_at": user.LastLoginAt,
		"password_changed_at": user.PasswordChangedAt,
	}
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	unvalid_chars_message := "%s must not contain any of the following characters: space, /, \\, ?, %, *, :, |, \", <, >"

	if req.Username == "" {
		WriteError(w, http.StatusBadRequest, "username cannot be empty")
		return
	}

	if containsRestrictedChars(req.Username) {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf(unvalid_chars_message, "username"))
		return
	}

	if req.Email == "" {
		WriteError(w, http.StatusBadRequest, "email cannot be empty")
		return
	}

	if !isValidEmail(req.Email) {
		WriteError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	if len(req.Password) < 8 {
		WriteError(w, http.StatusBadRequest, "password must be at least 8 characters long")
		return
	}

	if containsRestrictedChars(req.Password) {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf(unvalid_chars_message, "password"))
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to hash password")
		log.Println("failed to hash password:", err)
		return
	}

	log.Println("Creating user:", req.Username, req.Email)
	
	user_id, err := s.Q.CreateUser(r.Context(), db.CreateUserParams{
		Email:    req.Email,
		Username: req.Username,
		PasswordHash: hash,
		PasswordAlgo: "argon2id",
	})

	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create user")
		log.Println("failed to create user:", err)
		return
	}
	
	user, err := s.Q.GetUserByID(r.Context(), user_id)

	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to retrieve created user")
		log.Println("failed to retrieve created user:", err)
		return
	}

	writeJSON(w, http.StatusCreated, formatDBUser(user))
}
