package handlers

import (
	"context"
	"debt-manager/internal/contextkeys"
	"debt-manager/internal/db"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func readAccessTokenFromCookie(r *http.Request) (string, error) {
	if c, err := r.Cookie("access_token"); err == nil {
		if c.Value != "" {
			return c.Value, nil
		}
	}

	return "", errors.New("no access token found")
}

func (s *Server) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := readAccessTokenFromCookie(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "missing or invalid access token")
			return
		}

		token, err := jwt.ParseWithClaims(
			tokenStr,
			&Claims{},
			func(token *jwt.Token) (any, error) {
				return s.HS256PrivateKey, nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)
		if err != nil || !token.Valid {
			log.Println("Error parsing token or invalid token:", err)
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		claims := token.Claims.(*Claims)

		if len(claims.SessionID ) == 0 || len(claims.UserID) == 0 {
			log.Println("Invalid claims in token")
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		s.Tx.WithTx(r.Context(), func(q *db.Queries) error {
			session, err := q.GetSessionByID(
				r.Context(),
				pgtype.UUID{Bytes: uuid.MustParse(claims.SessionID), Valid: true},
			)

			if err != nil || session.RevokedAt.Valid || time.Now().After(session.ExpiresAt.Time) {
				log.Println("Error getting session or invalid session:", err)
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return err
			}

			ctx := context.WithValue(r.Context(), contextkeys.UserID{}, uuid.MustParse(claims.UserID))
			ctx = context.WithValue(ctx, contextkeys.SessionID{}, uuid.MustParse(claims.SessionID))
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		})
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		sw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		fmt.Printf(
			"%s %s -> %d (%v)\n",
			r.Method,
			r.URL.Path,
			sw.status,
			time.Since(start),
		)
	})
}
