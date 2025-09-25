package handlers

import "debt-manager/internal/db"

type Server struct {
	Tx *db.TxRunner
	Q  *db.Queries
}
