package main

import (
	"context"
	"debt-manager/internal/db"
	http_ "debt-manager/internal/http"
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

	q := db.New(pool)
	server := &handlers.Server{Q: q}

	mux := http.NewServeMux()
	mux.Handle("POST /lists", http_.UserMiddleware(http.HandlerFunc(server.CreateList)))
	mux.Handle("POST /users", http.HandlerFunc(server.CreateUser))

	const addr = ":8080"
	log.Printf("Starting server on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
