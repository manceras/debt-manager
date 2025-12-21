package main

import (
	"debt-manager/internal/config"
	"debt-manager/internal/db"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	DSN := "postgres://" + cfg.MigrationsUser + ":" + cfg.MigrationsPassword + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName
	log.Println("Migrating database with DSN:", DSN)
	db.Migrate(DSN)
	log.Println("Migrated", DSN)
}
