package handlers

import (
	"bytes"
	"context"
	"debt-manager/internal/db"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type DivisionRequest struct {
	OweUserID uuid.UUID `json:"owe_user_id"`
	Amount    float64   `json:"amount"`
}

type DivisionResponse struct {
	ID        uuid.UUID `json:"id"`
	OweUserID uuid.UUID `json:"owe_user_id"`
	Amount    float64   `json:"amount"`
}

type PaymentRequest struct {
	Title       string            `json:"title"`
	Amount      float64           `json:"amount"`
	PhotoURL    *string           `json:"photo_url"`
	PayerUserID uuid.UUID         `json:"payer_user_id"`
	Divisions   []DivisionRequest `json:"divisions"`
}

type PaymentResponse struct {
	ID          uuid.UUID         `json:"id"`
	Title       string            `json:"title"`
	Amount      float64           `json:"amount"`
	PhotoURL    *string           `json:"photo_url,omitempty"`
	PayerUserID uuid.UUID         `json:"payer_user_id"`
	Divisions   []DivisionRequest `json:"divisions"`
	CreatedAt   string            `json:"created_at"`
	ListID      uuid.UUID         `json:"list_id"`
}

type TransactionResponse struct {
	From   uuid.UUID `json:"from"`
	To     uuid.UUID `json:"to"`
	Amount float64   `json:"amount"`
}

type DepositRequest struct {
	Amount      float64   `json:"amount"`
	PayerUserID uuid.UUID `json:"from"`
	PayeeUserID uuid.UUID `json:"to"`
}

func parseJSONStrict(r io.ReadCloser, dst any) error {
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if len(bytes.TrimSpace(b)) == 0 {
		return fmt.Errorf("empty body")
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	dec.UseNumber()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values in body")
		}
		return fmt.Errorf("trailing data: %w", err)
	}

	return nil
}


var validate = validator.New()

func parseJSON(r io.ReadCloser, dst any) error {
	defer r.Close()

	if err := parseJSONStrict(r, dst); err != nil {
		return err
	}
	if err := validate.Struct(dst); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	return nil
}

func numericFromFloat(f float64, precission int) pgtype.Numeric {
	value := math.Round(f * math.Pow10(precission))
	i := big.NewInt(int64(value))
	return pgtype.Numeric{Int: i, Exp: int32(-precission), Valid: true}
}

func floatFromNumeric(n pgtype.Numeric) (float64, error) {
	if !n.Valid {
		return 0, errors.New("invalid numeric value")
	}
	floatVal, err := n.Float64Value()
	if err != nil {
		return 0, err
	}

	return floatVal.Float64, nil
}

func (s *Server) CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	listIDStr := chi.URLParam(r, "list_id")
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}
	listPgID := pgtype.UUID{Bytes: listID, Valid: true}

	var req PaymentRequest
	if err := parseJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		payerPgID := pgtype.UUID{Bytes: req.PayerUserID, Valid: true}
		var photoURL pgtype.Text
		if req.PhotoURL != nil {
			photoURL = pgtype.Text{String: *req.PhotoURL, Valid: true}
		} else {
			photoURL = pgtype.Text{Valid: false}
		}

		payment, err := q.CreatePayment(ctx, db.CreatePaymentParams{
			Title:       pgtype.Text{String: req.Title, Valid: true},
			PayerUserID: payerPgID,
			Amount:      numericFromFloat(req.Amount, 2),
			PhotoUrl:    photoURL,
			ListID:      listPgID,
		})

		if err != nil {
			log.Println("Error creating payment:", err)
			writeError(w, http.StatusInternalServerError, "failed to create payment")
			return err
		}

		divisionsTotal := 0.0
		for _, division := range req.Divisions {
			divisionsTotal += division.Amount
			owePgID := pgtype.UUID{Bytes: division.OweUserID, Valid: true}
			_, err := q.CreateDivision(ctx, db.CreateDivisionParams{
				PaymentID: payment.ID,
				OweUserID: owePgID,
				Amount:    numericFromFloat(division.Amount, 2),
			})
			if err != nil {
				log.Println("Error creating division:", err)
				writeError(w, http.StatusInternalServerError, "failed to create division")
				return err
			}
		}

		if req.Amount != divisionsTotal {
			log.Println("Error: payment amount does not match divisions total")
			writeError(w, http.StatusBadRequest, fmt.Sprintf("payment amount (%.2f) does not match divisions total (%.2f)", req.Amount, divisionsTotal))
			return errors.New("payment amount does not match divisions total")
		}

		writeJSON(w, http.StatusCreated, PaymentResponse{
			ID:          payment.ID.Bytes,
			Title:       payment.Title.String,
			Amount:      req.Amount,
			PhotoURL:    req.PhotoURL,
			PayerUserID: req.PayerUserID,
			Divisions:   req.Divisions,
			CreatedAt:   payment.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			ListID:      listID,
		})

		return nil
	})
}

func (s *Server) GetAllPaymentsForList(w http.ResponseWriter, r *http.Request) {
	listIDStr := chi.URLParam(r, "list_id")
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}

	pgListID := pgtype.UUID{Bytes: listID, Valid: true}
	err = s.Tx.WithCtxUserTx(r.Context(), func(q *db.Queries) error {
		payments, err := q.GetAllPaymentsForList(r.Context(), pgListID)
		if err != nil {
			log.Println("Error fetching payments:", err)
			writeError(w, http.StatusInternalServerError, "failed to fetch payments")
			return err
		}

		var resp []PaymentResponse
		for _, p := range payments {
			var photoURL *string
			if p.PhotoUrl.Valid {
				photoURL = &p.PhotoUrl.String
			}

			divisions, err := q.GetDivisionsByPaymentID(r.Context(), p.ID)
			if err != nil {
				log.Println("Error fetching divisions:", err)
				writeError(w, http.StatusInternalServerError, "failed to fetch divisions")
				return err
			}

			var divisionResponses []DivisionRequest
			for _, d := range divisions {
				amountFloat, err := floatFromNumeric(d.Amount)
				if err != nil {
					log.Println("Error converting amount:", err)
					writeError(w, http.StatusInternalServerError, "failed to convert amount")
					return err
				}
				divisionResponses = append(divisionResponses, DivisionRequest{
					OweUserID: d.OweUserID.Bytes,
					Amount:    amountFloat,
				})
			}

			amountFloat, err := floatFromNumeric(p.Amount)
			if err != nil {
				log.Println("Error converting amount:", err)
				writeError(w, http.StatusInternalServerError, "failed to convert amount")
				return err
			}
			resp = append(resp, PaymentResponse{
				ID:          p.ID.Bytes,
				Title:       p.Title.String,
				Amount:      amountFloat,
				PhotoURL:    photoURL,
				PayerUserID: p.PayerUserID.Bytes,
				Divisions:   divisionResponses,
				CreatedAt:   p.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
				ListID:      p.ListID.Bytes,
			})
		}

		writeJSON(w, http.StatusOK, resp)

		return nil
	})
}

func (s *Server) DeletePaymentByID(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := chi.URLParam(r, "payment_id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment ID")
		return
	}
	pgPaymentID := pgtype.UUID{Bytes: paymentID, Valid: true}

	err = s.Tx.WithCtxUserTx(r.Context(), func(q *db.Queries) error {
		err := q.DeletePaymentByID(r.Context(), pgPaymentID)
		if err != nil {
			log.Println("Error deleting payment:", err)
			writeError(w, http.StatusInternalServerError, "failed to delete payment")
			return err
		}
		return nil
	})

	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getBalancesFromPayments(p []db.Payment, d []db.Division, dep []db.Deposit) (map[uuid.UUID]float64, error) {
	balances := make(map[uuid.UUID]float64)

	for _, payment := range p {
		paymentAmount, err := floatFromNumeric(payment.Amount)
		if err != nil {
			return nil, err
		}
		balances[payment.PayerUserID.Bytes] += paymentAmount
		balances[payment.PayerUserID.Bytes] = math.Round(balances[payment.PayerUserID.Bytes]*100) / 100
	}

	for _, deposit := range dep {
		depositAmount, err := floatFromNumeric(deposit.Amount)
		if err != nil {
			return nil, err
		}
		balances[deposit.PayeeUserID.Bytes] -= depositAmount
		balances[deposit.PayerUserID.Bytes] += depositAmount
		balances[deposit.PayeeUserID.Bytes] = math.Round(balances[deposit.PayeeUserID.Bytes]*100) / 100
		balances[deposit.PayerUserID.Bytes] = math.Round(balances[deposit.PayerUserID.Bytes]*100) / 100
	}

	for _, division := range d {
		divisionAmount, err := floatFromNumeric(division.Amount)
		if err != nil {
			return nil, err
		}
		balances[division.OweUserID.Bytes] -= divisionAmount
		balances[division.OweUserID.Bytes] = math.Round(balances[division.OweUserID.Bytes]*100) / 100
	}

	return balances, nil
}

func getBalancesByListID(q *db.Queries, ctx context.Context, listID pgtype.UUID) (map[uuid.UUID]float64, error) {
	payments, err := q.GetAllPaymentsForList(ctx, listID)
	if err != nil {
		return nil, err
	}
	var divisions []db.Division
	for _, payment := range payments {
		paymentDivisions, err := q.GetDivisionsByPaymentID(ctx, payment.ID)
		if err != nil {
			return nil, err
		}
		divisions = append(divisions, paymentDivisions...)
	}
	deposits, err := q.GetAllDepositsForListID(ctx, listID)
	return getBalancesFromPayments(payments, divisions, deposits)
}

func (s *Server) GetNetBalances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	listIDStr := chi.URLParam(r, "list_id")
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}
	pgListID := pgtype.UUID{Bytes: listID, Valid: true}

	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		balances, err := getBalancesByListID(q, ctx, pgListID)
		if err != nil {
			log.Println("Error fetching net balances:", err)
			writeError(w, http.StatusInternalServerError, "failed to fetch net balances")
			return err
		}

		writeJSON(w, http.StatusOK, balances)
		return nil
	})
}

func (s *Server) GetSugestedTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	listIDStr := chi.URLParam(r, "list_id")
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}
	pgListID := pgtype.UUID{Bytes: listID, Valid: true}

	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		balances, err := getBalancesByListID(q, ctx, pgListID)
		if err != nil {
			log.Println("Error fetching net balances:", err)
			writeError(w, http.StatusInternalServerError, "failed to fetch net balances")
			return err
		}

		var creditors []struct {
			UserID uuid.UUID
			Amount float64
		}
		var debtors []struct {
			UserID uuid.UUID
			Amount float64
		}

		for userID, balance := range balances {
			if balance > 0 {
				creditors = append(creditors, struct {
					UserID uuid.UUID
					Amount float64
				}{UserID: userID, Amount: balance})
			} else if balance < 0 {
				debtors = append(debtors, struct {
					UserID uuid.UUID
					Amount float64
				}{UserID: userID, Amount: -balance})
			}
		}

		var transactions []TransactionResponse
		i, j := 0, 0
		for i < len(debtors) && j < len(creditors) {
			debtor := &debtors[i]
			creditor := &creditors[j]

			minAmount := math.Min(debtor.Amount, creditor.Amount)
			if minAmount > 0 {
				transactions = append(transactions, TransactionResponse{
					From:   debtor.UserID,
					To:     creditor.UserID,
					Amount: math.Round(minAmount*100) / 100,
				})

				debtor.Amount -= minAmount
				creditor.Amount -= minAmount
			}

			if debtor.Amount == 0 {
				i++
			}
			if creditor.Amount == 0 {
				j++
			}
		}
		writeJSON(w, http.StatusOK, transactions)

		return nil
	})
}

func (s *Server) CreateDeposit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	listIDStr := chi.URLParam(r, "list_id")
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list ID")
		return
	}
	pgListID := pgtype.UUID{Bytes: listID, Valid: true}

	var req DepositRequest
	if err := parseJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	log.Println("Deposit request:", req)

	err = s.Tx.WithCtxUserTx(ctx, func(q *db.Queries) error {
		deposit, err := q.CreateDeposit(ctx, db.CreateDepositParams{
			ListID:      pgListID,
			Amount:      numericFromFloat(req.Amount, 2),
			PayerUserID: pgtype.UUID{Bytes: req.PayerUserID, Valid: true},
			PayeeUserID: pgtype.UUID{Bytes: req.PayeeUserID, Valid: true},
		})
		if err != nil {
			log.Println("Error creating deposit:", err)
			writeError(w, http.StatusInternalServerError, "failed to create deposit")
			return err
		}
		writeJSON(w, http.StatusCreated, TransactionResponse{
			From:   deposit.PayerUserID.Bytes,
			To:     deposit.PayeeUserID.Bytes,
			Amount: req.Amount,
		})

		return nil
	})
}
