package main

import (
	"log"
	"os"
	"debt-manager/internal/db"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	db.Migrate(dsn)
}
