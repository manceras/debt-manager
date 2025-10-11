package config

import (
	"log"
	"net/url"
	"os"
)

type Config struct {
	Port            	 	string
	DBHost						 	string
	DBPort						 	string
	DBName 						 	string
	DBUser						 	string
	DBPassword				 	string
	MigrationsUser		 	string
	MigrationsPassword 	string
	JWTSecretKey			 	string
	BaseURL 						*url.URL
}

func baseURL(protocol, host, port string) string {
	if (protocol == "http" && port == "80") || (protocol == "https" && port == "443") {
		return protocol + "://" + host
	}
	return protocol + "://" + host + ":" + port
}

func Load() (Config, error) {
	stringURL := baseURL(
		getenv("PROTOCOL", "http"),
		getenv("HOST", "localhost"),
		getenv("PORT", "8080"),
	)

	baseURL, err := url.Parse(stringURL)
	if err != nil {
		log.Fatalf("Invalid BASE_URL: %v", err)
	}

	cfg := Config{
		Port: getenv("PORT", "8080"),
		DBHost: getenv("DB_HOST"),
		DBPort: getenv("DB_PORT"),
		DBName: getenv("DB_NAME"),
		DBUser: getenv("DB_USER"),
		DBPassword: getenv("DB_PASSWORD"),
		MigrationsUser: getenv("MIGRATIONS_USER"),
		MigrationsPassword: getenv("MIGRATIONS_PASSWORD"),
		JWTSecretKey: getenv("JWT_SECRET_KEY"),
		BaseURL: baseURL,
	}
	return cfg, nil
}

func getenv(k string, def ...string) string {
	d := ""
	if len(def) > 0 {
		d = def[0]
	}

	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
