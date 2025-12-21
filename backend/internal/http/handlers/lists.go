package handlers

import (
	"database/sql"
	"debt-manager/internal/contextkeys"
	"debt-manager/internal/db"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateListRequest struct {
	Title    string
	Currency string
}

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyJPY Currency = "JPY"
	CurrencyCNY Currency = "CNY"
)

type ListResponse struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Currency string    `json:"currency"`
	CreatedAt string    `json:"created_at"`
}

func (c Currency) Valid() bool {
	switch c {
	case CurrencyUSD, CurrencyEUR, CurrencyGBP, CurrencyJPY, CurrencyCNY:
		return true
	default:
		return false
	}
}

func (s *Server) CreateList(w http.ResponseWriter, r *http.Request) {
	var req CreateListRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}

	if !Currency(req.Currency).Valid() {
		writeError(w, http.StatusBadRequest, "currency not valid")
		return
	}

	ctx := r.Context()
	err := s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		newListID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		err := q.CreateList(r.Context(), db.CreateListParams{
			ID:       newListID,
			Title:    req.Title,
			Currency: db.Currency(req.Currency),
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create list")
			log.Println("failed to create list:", err)
			return err
		}

		var userID = ctx.Value(contextkeys.UserID{}).(uuid.UUID)

		_, err = q.CreateUserListRelation(r.Context(), db.CreateUserListRelationParams{
			UserID: pgtype.UUID{Bytes: userID, Valid: true},
			ListID: newListID,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create user-list relation")
			log.Println("failed to create user-list relation:", err)
			return err
		}

		list, err := q.GetListByID(r.Context(), newListID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve created list")
			log.Println("failed to retrieve created list:", err)
			return err
		}

		writeJSON(w, http.StatusOK, ListResponse{
			ID:        list.ID.Bytes,
			Title:     list.Title,
			Currency:  string(list.Currency),
			CreatedAt: list.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		})
		return nil
	})

	if err != nil {
		log.Println("transaction failed:", err)
		return
	}

}

func (s *Server) GetLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var lists []db.List
	err := s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		var err error
		lists, err = q.GetAllLists(ctx)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve lists")
			log.Println("failed to retrieve lists:", err)
			return err
		}
		return nil
	})

	if err != nil {
		log.Println("transaction failed:", err)
		return
	}

	responses := make([]ListResponse, len(lists))
	for i, list := range lists {
		responses[i] = ListResponse{
			ID:        list.ID.Bytes,
			Title:     list.Title,
			Currency:  string(list.Currency),
			CreatedAt: list.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	writeJSON(w, http.StatusOK, responses)
}

func (s *Server) GetListByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "list_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}
	PGID := pgtype.UUID{Bytes: id, Valid: true}
	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		list, err := q.GetListByID(ctx, PGID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "list not found")
				return nil
			}
			writeError(w, http.StatusInternalServerError, "failed to retrieve list")
			log.Println("failed to retrieve list:", err)
			return err
		}

		writeJSON(w, http.StatusOK, ListResponse{
			ID:        list.ID.Bytes,
			Title:     list.Title,
			Currency:  string(list.Currency),
			CreatedAt: list.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		})
		return nil
	})
	if err != nil {
		log.Println("transaction failed:", err)
		return
	}
}

type UpdateListRequest struct {
	Title 	 *string   `json:"title,omitempty"`
	Currency *Currency `json:"currency,omitempty"`
}

func (s *Server) UpdateList(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "list_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}

	var req UpdateListRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Title != nil && *req.Title == "" {
		writeError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}

	if req.Currency != nil && !req.Currency.Valid() {
		writeError(w, http.StatusBadRequest, "currency not valid")
		return
	}

	var title pgtype.Text
	if req.Title != nil {
		title = pgtype.Text{String: *req.Title, Valid: true}
	} else {
		title = pgtype.Text{Valid: false}
	}

	var currency db.NullCurrency
	if req.Currency != nil {
		currency = db.NullCurrency{Currency: db.Currency(*req.Currency), Valid: true}
	} else {
		currency = db.NullCurrency{Valid: false}
	}

	PGID := pgtype.UUID{Bytes: id, Valid: true}
	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		err := q.UpdateList(ctx, db.UpdateListParams{
			ID:       PGID,
			Title:  	title,
			Currency: currency,
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "list not found")
				return nil
			}
			writeError(w, http.StatusInternalServerError, "failed to retrieve list")
			log.Println("failed to retrieve list:", err)
			return err
		}

		list, err := q.GetListByID(ctx, PGID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve list")
			log.Println("failed to retrieve list:", err)
			return err
		}

		writeJSON(w, http.StatusOK, ListResponse{
			ID:        list.ID.Bytes,
			Title:     list.Title,
			Currency:  string(list.Currency),
			CreatedAt: list.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		})
		return nil
	})
}

func (s *Server) DeleteList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "list_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}
	PGID := pgtype.UUID{Bytes: id, Valid: true}
	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		_, err := q.DeleteList(ctx, PGID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "list not found")
				return nil
			}
			writeError(w, http.StatusInternalServerError, "failed to delete list")
			log.Println("failed to delete list:", err)
			return err
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	})
	if err != nil {
		log.Println("transaction failed:", err)
		return
	}
}
