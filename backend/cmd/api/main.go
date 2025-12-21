package main

import (
	"context"
	app "debt-manager/internal"
	"debt-manager/internal/config"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	DBDSN := "postgres://" + cfg.DBUser + ":" + cfg.DBPassword + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName
	a, err := app.New(ctx, DBDSN, []byte(cfg.JWTSecretKey))
	if err != nil {
		log.Fatal("cannot create app:", err)
	}
	defer a.Close()

	log.Printf("starting server at :%s...", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, a.Mux)

}
