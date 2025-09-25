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
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Title == "" {
		WriteError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}

	if !Currency(req.Currency).Valid() {
		WriteError(w, http.StatusBadRequest, "currency not valid")
		return
	}

	list, err := s.Q.CreateList(r.Context(), db.CreateListParams{
		Title:    req.Title,
		Currency: db.Currency(req.Currency),
	})

	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create list")
		log.Println("failed to create list:", err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}
