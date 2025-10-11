package handlers

import (
	"crypto/rand"
	"database/sql"
	"debt-manager/internal/config"
	"debt-manager/internal/contextkeys"
	"debt-manager/internal/db"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func generateInvitationHash() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return "INV" + hex.EncodeToString(b), nil
}

type InvitationResponse struct {
	ID        uuid.UUID `json:"id"`
	Hash      string    `json:"hash"`
	CreatedAt string    `json:"created_at"`
	ExpiresAt string    `json:"expires_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	InvitedBy *string   `json:"invited_by,omitempty"`
	ListTitle *string   `json:"list_title,omitempty"`
	RevokedAt *string   `json:"revoked_at,omitempty"`
}

func generateInvitationLink(hash string) (string, error) {
	cfg, err := config.Load()
	if err != nil {
		log.Println("failed to load config:", err)
		return "", err
	}

	base := *cfg.BaseURL
	base.Path += "/invitations/" + hash

	return base.String(), nil
}

func (s *Server) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	listId, err := uuid.Parse(chi.URLParam(r, "list_id"))
	PGListId := pgtype.UUID{Bytes: listId, Valid: true}
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}

	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		_, err := q.GetListByID(ctx, PGListId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "list not found")
				return nil
			}
			writeError(w, http.StatusInternalServerError, "failed to retrieve list")
			log.Println("failed to retrieve list:", err)
			return err
		}

		invitationHash, err := generateInvitationHash()
		if err != nil {
			log.Println("failed to generate invitation hash:", err)
			writeError(w, http.StatusInternalServerError, "failed to generate invitation hash")
			return err
		}

		createdBy := ctx.Value(contextkeys.UserID{}).(uuid.UUID)
		_, err = q.CreateInvitation(ctx, db.CreateInvitationParams{
			InvitedToListID: PGListId,
			Hash:            invitationHash,
			CreatedBy:       pgtype.UUID{Bytes: createdBy, Valid: true},
			ExpiresAt:       pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true},
		})
		if err != nil {
			log.Println("failed to create invitation:", err)
			writeError(w, http.StatusInternalServerError, "failed to create invitation")
			return err
		}

		invitationLink, err := generateInvitationLink(invitationHash)
		if err != nil {
			log.Println("failed to generate invitation link:", err)
			writeError(w, http.StatusInternalServerError, "failed to generate invitation link")
			return err
		}
		writeJSON(w, http.StatusCreated, map[string]string{
			"invitation_link": invitationLink,
		})
		return nil
	})
	if err != nil {
		log.Println("transaction error:", err)
	}
}

func (s *Server) GetAllInvitationsForList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list_id, err := uuid.Parse(chi.URLParam(r, "list_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}

	PGListId := pgtype.UUID{Bytes: list_id, Valid: true}
	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		invitations, err := q.GetAllInvitationsForList(ctx, PGListId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "list not found")
				return nil
			}
			writeError(w, http.StatusInternalServerError, "failed to retrieve invitations")
			log.Println("failed to retrieve invitations:", err)
			return err
		}


		responses := make([]InvitationResponse, len(invitations))
		for i, invitation := range invitations {
			invitedByUser, err := q.GetUserByID(ctx, invitation.CreatedBy)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to retrieve user")
				log.Println("failed to retrieve user:", err)
			}
			listTitle, err := q.GetListByID(ctx, invitation.InvitedToListID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to retrieve list")
				log.Println("failed to retrieve list:", err)
			}

			responses[i] = InvitationResponse{
				ID:        invitation.ID.Bytes,
				Hash:      invitation.Hash,
				CreatedAt: invitation.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
				ExpiresAt: invitation.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00"),
				CreatedBy: invitation.CreatedBy.Bytes,
				InvitedBy: &invitedByUser.Username,
				ListTitle: &listTitle.Title,
			}
		}

		writeJSON(w, http.StatusOK, responses)
		return nil
	})
}

func (s *Server) GetInvitationByHash(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		writeError(w, http.StatusBadRequest, "invalid invitation hash")
		return
	}

	err := s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		invitation, err := q.GetInvitationByHash(ctx, hash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "invitation not found")
				return nil
			}
			writeError(w, http.StatusInternalServerError, "failed to retrieve invitation")
			log.Println("failed to retrieve invitation:", err)
			return err
		}

		list, err := q.GetListByID(ctx, invitation.InvitedToListID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve list")
			log.Println("failed to retrieve list:", err)
		}

		invitedByUser, err := q.GetUserByID(ctx, invitation.CreatedBy)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve user")
			log.Println("failed to retrieve user:", err)
		}

		revokedAt := invitation.RevokedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		response := InvitationResponse{
			ID:        invitation.ID.Bytes,
			Hash:      invitation.Hash,
			CreatedAt: invitation.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			ExpiresAt: invitation.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy: invitation.CreatedBy.Bytes,
			InvitedBy: &invitedByUser.Username,
			ListTitle: &list.Title,
			RevokedAt: &revokedAt,
		}

		writeJSON(w, http.StatusOK, response)
		return nil
	})
	if err != nil {
		log.Println("transaction error:", err)
	}
}

func (s *Server) RevokeInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	invitationId, err := uuid.Parse(chi.URLParam(r, "invitation_id"))
	PGInvitationId := pgtype.UUID{Bytes: invitationId, Valid: true}
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invitation ID")
		return
	}

	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		err := q.RevokeInvitationByID(ctx, PGInvitationId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "invitation not found")
				return nil
			}
			writeError(w, http.StatusInternalServerError, "failed to delete invitation")
			log.Println("failed to delete invitation:", err)
			return err
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	})
	if err != nil {
		log.Println("transaction error:", err)
	}
}

func (s *Server) GetInvitationByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	invitationId, err := uuid.Parse(chi.URLParam(r, "invitation_id"))
	PGInvitationId := pgtype.UUID{Bytes: invitationId, Valid: true}
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invitation ID")
		return
	}
	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		invitation, err := q.GetInvitationByID(ctx, PGInvitationId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "invitation not found")
				return nil
			}
			writeError(w, http.StatusInternalServerError, "failed to retrieve invitation")
			log.Println("failed to retrieve invitation:", err)
			return err
		}

		list, err := q.GetListByID(ctx, invitation.InvitedToListID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve list")
			log.Println("failed to retrieve list:", err)
		}

		invitedByUser, err := q.GetUserByID(ctx, invitation.CreatedBy)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve user")
			log.Println("failed to retrieve user:", err)
		}

		revokedAt := invitation.RevokedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		response := InvitationResponse{
			ID:        invitation.ID.Bytes,
			Hash:      invitation.Hash,
			CreatedAt: invitation.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			ExpiresAt: invitation.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy: invitation.CreatedBy.Bytes,
			InvitedBy: &invitedByUser.Username,
			ListTitle: &list.Title,
			RevokedAt: &revokedAt,
		}

		writeJSON(w, http.StatusOK, response)
		return nil
	})
}
