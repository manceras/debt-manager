package handlers

import (
	"debt-manager/internal/contextkeys"
	"debt-manager/internal/db"
	"log"
	"net/http"

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

		writeJSON(w, http.StatusOK, list)
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

	writeJSON(w, http.StatusOK, lists)
}

func (s *Server) GetListByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var list db.List
	err := s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		list, err = q.GetListByID(ctx)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve list")
			log.Println("failed to retrieve list:", err)
			return err
		}
		return nil
}
