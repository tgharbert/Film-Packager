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

func PoolConnect() *pgxpool.Pool {
	dbURL := os.Getenv("RDS_URL")
	if dbURL == "" {
		panic("database url not found in environment")
	}

	fmt.Println("dbURL: ", dbURL)
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		fmt.Println("error parsing config: ", err)
		panic(err)
	}

	// later need to work on pool settings
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		fmt.Println("error creating pool: ", err)
		panic(err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		fmt.Println("error pinging pool: ", err)
		panic(err)
	}

	return pool
}
