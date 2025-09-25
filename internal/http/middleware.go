package http

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKeyUserID struct{}

func UserIDFromContext(ctx context.Context) uuid.UUID {
	val := ctx.Value(ctxKeyUserID{})
	if id, ok := val.(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("X-User-ID")
		if header == "" {
			http.Error(w, "missing X-User-ID header", http.StatusBadRequest)
			return
		}
		userID, err := uuid.Parse(header)
		if err != nil {
			http.Error(w, "invalid X-User-ID header", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), ctxKeyUserID{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
