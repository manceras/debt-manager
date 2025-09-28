package handlers

import (
	"context"
	"debt-manager/internal/contextkeys"
	"debt-manager/internal/db"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func readAccessTokenFromCookieOrHeader(r *http.Request) (string, error) {
	if h := r.Header.Get("Authorization"); h != "" {
		if len(h) > 7 && h[:7] == "Bearer " {
			t := strings.TrimSpace(h[7:])
			if t != "" {
				return t, nil
			}
		}
	}
	if c, err := r.Cookie("access_token"); err == nil {
		if c.Value != "" {
			return c.Value, nil
		}
	}

	return "", errors.New("no access token found")
}

func (s *Server) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := readAccessTokenFromCookieOrHeader(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "missing or invalid access token")
			return
		}

		token, err := jwt.ParseWithClaims(
			tokenStr,
			&Claims{},
			func(token *jwt.Token) (interface{}, error) {
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

			ctx := context.WithValue(r.Context(), contextkeys.UserID{}, claims.UserID)
			ctx = context.WithValue(ctx, contextkeys.SessionID{}, claims.SessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		})
	})
}
