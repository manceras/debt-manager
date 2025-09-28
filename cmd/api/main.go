package main

import (
	"context"
	"debt-manager/internal/db"
	"debt-manager/internal/http/handlers"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)
func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	migrations_dsn := os.Getenv("MIGRATIONS_DSN")
	dsn := os.Getenv("DATABASE_DSN")
	db.Migrate(migrations_dsn)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()


	server := &handlers.Server{
		HS256PrivateKey: []byte(os.Getenv("HS256_PRIVATE_KEY")),
		Tx: db.NewTxRunner(pool),
	}

	mux := http.NewServeMux()
	mux.Handle("POST /lists", server.Auth(http.HandlerFunc(server.CreateList)))

	mux.Handle("POST /auth/signup", http.HandlerFunc(server.SignUp))
	mux.Handle("POST /auth/login", http.HandlerFunc(server.Login))


	const addr = ":8080"
	log.Printf("Starting server on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
