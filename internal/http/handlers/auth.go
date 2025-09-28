package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"debt-manager/internal/db"
	"encoding/base64"
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

func createRefreshToken() (raw string, hash []byte, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil, err
	}
	raw = base64.URLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(raw))
	return raw, h[:], nil
}

func (s *Server) createSession(user db.AppUsersSafe, w http.ResponseWriter, r *http.Request) {
	s.Tx.WithTx(r.Context(), func(q *db.Queries) error {
		session, err := q.CreateSession(r.Context(), db.CreateSessionParams{
			UserID:    user.ID,
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(365 * 24 * time.Hour), Valid: true},
			UserAgent: pgtype.Text{String: r.UserAgent(), Valid: true},
			Ip: pgtype.Text{String: r.RemoteAddr, Valid: true},
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create session")
			log.Println("failed to create session:", err)
			return err
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
			return err
		}

		rtRaw, rtHash, err := createRefreshToken()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create refresh token")
			log.Println("failed to create refresh token:", err)
			return err
		}

		expiresAt := time.Now().Add(45 * 24 * time.Hour) // 45 days
		_, err = q.CreateRefreshToken(r.Context(), db.CreateRefreshTokenParams{
			SessionID: session.ID,
			TokenHash: rtHash,
			ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to store refresh token")
			log.Println("failed to store refresh token:", err)
			return err
		}

		setCookie(w, "access_token", signed, accessTTL)
		setCookie(w, "refresh_token", rtRaw, time.Until(expiresAt))

		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

func (s *Server) SignUp(w http.ResponseWriter, r *http.Request) {
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

	s.Tx.WithTx(r.Context(), func(q *db.Queries) error {
		user_id, err := q.CreateUser(r.Context(), db.CreateUserParams{
			Email: req.Email,
			Username: req.Username,
			PasswordHash: hash,
			PasswordAlgo: "argon2id",
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create user")
			log.Println("failed to create user:", err)
			return err
		}

		user, err := q.GetUserByID(r.Context(), user_id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve created user")
			log.Println("failed to retrieve created user:", err)
			return err
		}

		s.createSession(user, w, r)
		return nil
	})
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

	s.Tx.WithTx(r.Context(), func(q *db.Queries) error {
		user, err := q.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			log.Println("failed to get user by email:", err)
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return err
		}

		loginSecrets, err := q.GetLoginSecretsByEmail(r.Context(), user.Email)
		if err != nil {
			log.Println("failed to get login secrets by email:", err)
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return err
		}

		password_ok, err := VerifyPassword(req.Password, loginSecrets.PasswordHash.String)
		if err != nil || !password_ok {
			log.Println("failed to verify password:", err)
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return err
		}

		if err := q.UpdateUserLastLogin(r.Context(), user.ID); err != nil {
			log.Println("failed to update last login:", err)
			return err
		}

		s.createSession(user, w, r)
		return nil
	})
}
