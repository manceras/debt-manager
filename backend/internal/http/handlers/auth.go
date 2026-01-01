package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"debt-manager/internal/contextkeys"
	"debt-manager/internal/db"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserRequest struct {
	Username string	`json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

const(
		expiration_time = 45 * 24 * time.Hour // 45 days
		atTTL = 15 * time.Minute // 15 minutes
)

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

func (s *Server) makeAccessToken(claims *Claims) (string, error) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		signed, err := token.SignedString(s.HS256PrivateKey)
		if err != nil {
			return "", err
		}
		return signed, nil
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

func (s *Server) createSession(user db.AppUsersSafe, w http.ResponseWriter, r *http.Request, q *db.Queries) error {
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

	rtRaw, rtHash, err := createRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create refresh token")
		log.Println("failed to create refresh token:", err)
		return err
	}

	expiresAt := time.Now().Add(expiration_time)
	_, err = q.CreateRefreshToken(r.Context(), db.CreateRefreshTokenParams{
		ID: 			pgtype.UUID{Bytes: uuid.New(), Valid: true},
		SessionID: session.ID,
		TokenHash: rtHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store refresh token")
		log.Println("failed to store refresh token:", err)
		return err
	}

	signed, err := s.makeAccessToken(&Claims{
		SessionID: session.ID.String(),
		UserID:    user.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(atTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "debt-manager",
			Subject:   fmt.Sprint(user.ID),
		},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create access token")
		log.Println("failed to create access token:", err)
	}

	setCookie(w, "refresh_token", rtRaw, time.Until(expiresAt), "/")
	setCookie(w, "access_token", signed, atTTL, "/")

	w.WriteHeader(http.StatusNoContent)

	return nil
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

	err = s.Tx.WithTx(r.Context(), func(q *db.Queries) error {
		user_id, err := q.CreateUser(r.Context(), db.CreateUserParams{
			Email: req.Email,
			Username: req.Username,
			PasswordHash: hash,
			PasswordAlgo: "argon2id",
		})
		if err != nil {
			log.Println("failed to create user:", err)
			if strings.Contains(err.Error(), "email_already_registered") {
				writeError(w, http.StatusConflict, "email or username are already registered")
				return nil
			}
			writeError(w, http.StatusInternalServerError, "failed to create user")
			return err
		}

		user, err := q.GetUserByID(r.Context(), user_id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve created user")
			log.Println("failed to retrieve created user:", err)
			return err
		}

		s.createSession(user, w, r, q)
		return nil
	})

	if err != nil {
		log.Println("transaction failed during signup:", err)
		return
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	SessionID string `json:"sessionId"`
	UserID    string `json:"userId"`
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

		s.createSession(user, w, r, q)
		return nil
	})
}

func (s *Server) Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	hash := sha256.Sum256([]byte(c.Value))

	ctx := r.Context()
	err = s.Tx.WithTx(ctx, func(q *db.Queries) error {
		row, err := q.AuthRefreshLookup(ctx, hash[:])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid refresh token")
			log.Println("failed to lookup refresh token:", err)
			return err
		}

		if row.SessionRevokedAt.Valid {
			writeError(w, http.StatusUnauthorized, "session revoked")
			return nil
		}

		if row.RtRevokedAt.Valid || row.RtExpiresAt.Time.Before(time.Now()) {
			writeError(w, http.StatusUnauthorized, "refresh token revoked")
			return nil
		}

		if row.RtReplacedByID.Valid {
			q.RevokeWholeSession(ctx, row.SessionID)
			clearCookie(w, "access_token")
			writeError(w, http.StatusUnauthorized, "refresh token reused")
			return errors.New("refresh token reused")
		}

		return nil
	})

	if err != nil {
		return
	}

	err = s.Tx.WithTx(ctx, func(q *db.Queries) error {
		row, err := q.AuthRefreshLookup(ctx, hash[:])
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get session details")
			log.Println("failed to get session details:", err)
			return err
		}
		expire_at := time.Now().Add(expiration_time)
		if row.MaxExpiresAt.Valid && row.MaxExpiresAt.Time.Before(expire_at) {
			expire_at = row.MaxExpiresAt.Time
		}

		if expire_at.Before(time.Now()) {
			q.RevokeWholeSession(ctx, row.SessionID)
			clearCookie(w, "access_token")
			clearCookie(w, "refresh_token")
			writeError(w, http.StatusUnauthorized, "session expired")
			return nil;
		}

		new_rt_raw, new_rt_hash, err := createRefreshToken()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create refresh token")
			log.Println("failed to create refresh token:", err)
			return err
		}

		new_rt_id := uuid.New()
		affected, err := q.MarkOldTokenReplaced(ctx, db.MarkOldTokenReplacedParams{
			ID: row.RtID,
			ReplacedByID: pgtype.UUID{Bytes: new_rt_id, Valid: true},
		})
		if err != nil || affected == 0 {
			writeError(w, http.StatusInternalServerError, "failed to mark old refresh token as replaced")
			log.Println("failed to mark old refresh token as replaced:", err)
			return err
		}

		_, err = q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
			ID: pgtype.UUID{Bytes: new_rt_id, Valid: true},
			SessionID: row.SessionID,
			TokenHash: new_rt_hash,
			ParentID: row.RtID,
			ExpiresAt: pgtype.Timestamptz{
				Time:  expire_at,
				Valid: true,
			},
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to store refresh token")
			log.Println("failed to store refresh token:", err)
			return err
		}

		at, err := s.makeAccessToken(&Claims{
			SessionID: row.SessionID.String(),
			UserID:    row.UserID.String(),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "debt-manager",
				Subject:   fmt.Sprint(row.UserID),
			},
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create access token")
			log.Println("failed to create access token:", err)
			return nil
		}

		setCookie(w, "refresh_token", new_rt_raw, time.Until(expire_at), "/")
		setCookie(w, "access_token", at, atTTL, "/")

		w.WriteHeader(http.StatusNoContent)

		return nil
	})
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		var sessionID uuid.UUID = ctx.Value(contextkeys.SessionID{}).(uuid.UUID)

		err := q.RevokeWholeSession(ctx, pgtype.UUID{Bytes: sessionID, Valid: true})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to revoke sessions")
			log.Println("failed to revoke sessions:", err)
			return err
		}

		clearCookie(w, "access_token")
		clearCookie(w, "refresh_token")

		w.WriteHeader(http.StatusNoContent)
		return nil
	})

	if err != nil {
		log.Println("transaction failed during Logout:", err)
		return
	}
}

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		var userID uuid.UUID = ctx.Value(contextkeys.UserID{}).(uuid.UUID)
		var pgUserID = pgtype.UUID{Bytes: userID, Valid: true}
		
		user, err := q.GetUserByID(ctx, pgUserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get user info")
			log.Println("failed to get user info:", err)
			return err
		}

		resp := UserResponse{
			ID:        user.ID.Bytes,
			Email:     user.Email,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}

		writeJSON(w, http.StatusOK, resp)
		return nil
	})

	if err != nil {
		log.Println("transaction failed during Me:", err)
		return
	}
}
