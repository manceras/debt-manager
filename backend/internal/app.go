package internal

import (
	"context"
	"debt-manager/internal/db"
	"debt-manager/internal/http"
	"debt-manager/internal/http/handlers"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	DB     *pgxpool.Pool
	Server *handlers.Server
	Mux    *chi.Mux
}

func New(ctx context.Context, dsn string, jwtKey []byte) (*App, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	server := &handlers.Server{
		Tx:              db.NewTxRunner(pool),
		HS256PrivateKey: jwtKey,
	}

	mux := http.NewMux(server)

	return &App{
		DB:     pool,
		Server: server,
		Mux:    mux,
	}, nil
}

func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
	log.Println("App closed")
}
