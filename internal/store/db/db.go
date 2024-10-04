package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type User struct {
	name string
	email string
	role string
}

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

func CreateUser(c *pgx.Conn, name string, email string, password byte[]) (User, error) {
	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3)`
	_, err := c.Exec(context.Background(), query, name, email, password)
	var user User
	if err != nil {
		return user, fmt.Errorf("failed to insert user into users table: %v", err)
	}
	getNewUserQuery := `GET user FROM users WHERE email = VALUES($1)`
	rows, err := c.Query(context.Background(), getNewUserQuery, email)
	if err != nil {
		return user, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&user.email, &user.name, &user.role)
		if err != nil {
			return user, fmt.Errorf("error scanning row %v", err)
		}
	}
	if rows.Err() != nil {
		return user, rows.Err()
	}
	return user, nil
}
