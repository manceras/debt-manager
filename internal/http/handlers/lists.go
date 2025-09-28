package handlers

import (
	"debt-manager/internal/db"
	"log"
	"net/http"
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

	err := s.Tx.WithCtxUserTx(r.Context(), func(q *db.Queries) error {
		list, err := q.CreateList(r.Context(), db.CreateListParams{
			Title:    req.Title,
			Currency: db.Currency(req.Currency),
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create list")
			log.Println("failed to create list:", err)
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
