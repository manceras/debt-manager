package handlers

import (
	"debt-manager/internal/db"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5/pgtype"
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
		"id":                  user.ID,
		"username":            user.Username,
		"email":               user.Email,
		"created_at":          user.CreatedAt,
		"last_login_at":       user.LastLoginAt,
		"password_changed_at": user.PasswordChangedAt,
	}
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	unvalid_chars_message := "%s must not contain any of the following characters: space, /, \\, ?, %, *, :, |, \", <, >"

	if req.Username == "" {
		writeError(w, http.StatusBadRequest, "username cannot be empty")
		return
	}

	if containsRestrictedChars(req.Username) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf(unvalid_chars_message, "username"))
		return
	}

	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email cannot be empty")
		return
	}

	if !isValidEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters long")
		return
	}

	if containsRestrictedChars(req.Password) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf(unvalid_chars_message, "password"))
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		log.Println("failed to hash password:", err)
		return
	}

	log.Println("Creating user:", req.Username, req.Email)

	user_id, err := s.Tx.Q().CreateUser(r.Context(), db.CreateUserParams{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hash,
		PasswordAlgo: "argon2id",
	})

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		log.Println("failed to create user:", err)
		return
	}

	user, err := s.Tx.Q().GetUserByID(r.Context(), user_id)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve created user")
		log.Println("failed to retrieve created user:", err)
		return
	}

	writeJSON(w, http.StatusCreated, formatDBUser(user))
}

type LoginRequest struct {
	Email    string
	Password string
}

type Claims struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// Decode JSON body
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate input
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email cannot be empty")
		return
	}

	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password cannot be empty")
		return
	}

	user, err := s.Tx.Q().GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		log.Println("failed to get user by email:", err)
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	loginSecrets, err := s.Tx.Q().GetLoginSecretsByEmail(r.Context(), user.Email)
	if err != nil {
		log.Println("failed to get login secrets by email:", err)
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	password_ok, err := VerifyPassword(req.Password, loginSecrets.PasswordHash.String)
	if err != nil || !password_ok {
		log.Println("failed to verify password:", err)
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := s.Tx.Q().UpdateUserLastLogin(r.Context(), user.ID); err != nil {
		log.Println("failed to update last login:", err)
		return
	}

	// Create server-side session
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
	session, err := s.Tx.Q().CreateSession(r.Context(), db.CreateSessionParams{
		UserID:    user.ID,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
		UserAgent: pgtype.Text{String: r.UserAgent(), Valid: true},
		Ip: pgtype.Text{String: r.RemoteAddr, Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		log.Println("failed to create session:", err)
		return
	}

	accessTTL := 15 * time.Minute
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		SessionID: session.ID.String(),
		UserID:    user.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "debt-manager",
			Subject:   fmt.Sprint(user.ID),
		},
	})
	signed, err := token.SignedString(s.HS256PrivateKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to sign token")
		log.Println("failed to sign token:", err)
		return
	}

	setCookie(w, "access_token", signed, accessTTL)
	setCookie(w, "refresh_token", session.ID.String(), time.Until(expiresAt))

	w.WriteHeader(http.StatusNoContent)
}
