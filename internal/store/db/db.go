package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func Connect() *pgx.Conn {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error loading env file")
		panic(err)
	}
	conn, err := pgx.Connect(context.Background(), os.Getenv("DEV_DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to the database: %v\n", err)
		os.Exit(1)
	}
	return conn
}