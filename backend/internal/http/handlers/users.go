package handlers

import (
	"debt-manager/internal/contextkeys"
	"debt-manager/internal/db"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt string    `json:"created_at"`
	ItsYou    bool      `json:"its_you,omitempty"`
}

func (s *Server) GetUsersFromList(w http.ResponseWriter, r *http.Request) {
	listIDStr := chi.URLParam(r, "list_id")
	listID, err := uuid.Parse(listIDStr)
	listPgID := pgtype.UUID{Bytes: listID, Valid: true}
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}

	ctx := r.Context()
	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		users, err := q.GetUsersFromList(ctx, listPgID)
		if err != nil {
			log.Println("Error fetching users from list:", err)
			writeError(w, http.StatusInternalServerError, "failed to fetch users from list")
			return err
		}

		var resp []UserResponse
		for _, user := range users {
			var contextUserID uuid.UUID = ctx.Value(contextkeys.UserID{}).(uuid.UUID)
			itsYou := user.ID.Bytes == contextUserID
			resp = append(resp, UserResponse{
				ID:        user.ID.Bytes,
				Email:     user.Email,
				Username:  user.Username,
				CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
				ItsYou:    itsYou,
			})
		}
		writeJSON(w, http.StatusOK, resp)

		return nil
	})

}
