package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type User struct {
	Name string
	Email string
	Role string
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

func CreateUser(c *pgx.Conn, name string, email string, password string) (User, error) {
	query := `INSERT INTO users (name, email, password, role) VALUES ($1, $2, $3, $4)`
	role := "read"
	_, err := c.Exec(context.Background(), query, name, email, password, role)
	var user User
	if err != nil {
		return user, fmt.Errorf("failed to insert user into users table: %v", err)
	}
	// ERROR IS BELOW...
	getNewUserQuery := `SELECT email, name, role FROM users WHERE email = $1`
	rows, err := c.Query(context.Background(), getNewUserQuery, email)
	if err != nil {
		fmt.Println("HIT THE EEARLIER EERRRR")
		return user, err
	}
	defer rows.Close()

	if !rows.Next() {
		fmt.Println("No rows found for email:", email)
		return user, fmt.Errorf("no user found with email: %s", email)
	}

	// Scan the result into the user struct
err = rows.Scan(&user.Email, &user.Name, &user.Role)
if err != nil {
	fmt.Println("Error scanning row:", err)
	return user, fmt.Errorf("error scanning row: %v", err)
}

// Check for any errors in rows after the loop
if rows.Err() != nil {
	fmt.Println("Error after rows loop:", rows.Err())
	return user, rows.Err()
}

return user, nil

	// for rows.Next() {
	// 	err := rows.Scan(nil, &user.Name, &user.Email, nil, &user.Role)
	// 	if err != nil {
	// 		fmt.Println("SCANNING ERROR??")
	// 		return user, fmt.Errorf("error scanning row %v", err)
	// 	}
	// }
	// if rows.Err() != nil {
	// 	fmt.Println("ROWS	 ERROR??")
	// 	return user, rows.Err()
	// }
	// return user, nil
}
