package config

import "os"

type Config struct {
	Port            	 string
	DBHost						 string
	DBPort						 string
	DBName 						 string
	DBUser						 string
	DBPassword				 string
	MigrationsUser		 string
	MigrationsPassword string
	JWTSecretKey			 string
}

func Load() (Config, error) {
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
