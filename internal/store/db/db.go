package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func CheckPasswordHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// globally available db pool for connections
var DBPool *pgxpool.Pool

func PoolConnect() {
	dbURL := os.Getenv("DEV_DATABASE_URL")
	if dbURL == "" {
		fmt.Println("DEV_DATABASE_URL not found in environment")
		os.Exit(1)
	}

	// later need to work on pool settings
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	DBPool = pool
}

func GetPool() *pgxpool.Pool {
	return DBPool
}
