package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func CheckPasswordHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func PoolConnect() *pgxpool.Pool {
	var dbURL string
	if os.Getenv("ENV") == "DEV" {
		dbURL = os.Getenv("DEV_DATABASE_URL")
	} else {
		dbURL = os.Getenv("PROD_DB_URL")
	}

	if dbURL == "" {
		panic("database url not found in environment")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		panic(err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		panic(err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		panic(err)
	}

	return pool
}
