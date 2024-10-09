package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id int
	Name string
	Email string
	Role string
	Password string
}

type Org struct {
	Id string
	Name string
	Role string
}

func CheckPasswordHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
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
	role := "readonly"
	_, err := c.Exec(context.Background(), query, name, email, password, role)
	var user User
	if err != nil {
		return user, fmt.Errorf("failed to insert user into users table: %v", err)
	}
	getNewUserQuery := `SELECT email, name, role FROM users WHERE email = $1`
	rows, err := c.Query(context.Background(), getNewUserQuery, email)
	if err != nil {
		return user, err
	}
	defer rows.Close()
	if !rows.Next() {
		fmt.Println("No rows found for email:", email)
		return user, fmt.Errorf("no user found with email: %s", email)
	}
	err = rows.Scan(&user.Email, &user.Name, &user.Role)
	if err != nil {
		fmt.Println("Error scanning row:", err)
		return user, fmt.Errorf("error scanning row: %v", err)
	}
	if rows.Err() != nil {
		fmt.Println("Error after rows loop:", rows.Err())
		return user, rows.Err()
	}
	return user, nil
}

func GetUser(c *pgx.Conn, email string, password string) (User, error) {
	query := `SELECT id, email, name, role, password FROM users where email = $1`
	var user User
	rows, err := c.Query(context.Background(), query, email)
	if err != nil {
		return user, err
	}
	defer rows.Close()
	if !rows.Next() {
		fmt.Println("No rows found for email:", email)
		return user, fmt.Errorf("no user found with email: %s", email)
	}
	// var storedPassword string
	err = rows.Scan(&user.Id, &user.Email, &user.Name, &user.Role, &user.Password)
	if err != nil {
		fmt.Println("Error scanning row:", err)
		return user, fmt.Errorf("error scanning row: %v", err)
	}
	err = CheckPasswordHash(user.Password, password)
	if err != nil {
		user.Password = ""
		return user, fmt.Errorf("invalid password")
	}
	// if password != user.Password {
	// 	fmt.Println("WRONG PASSWORD - ADD LOGIC")
	// 	user.Email = "false"
	// 	return user, err
	// }
	if rows.Err() != nil {
		fmt.Println("Error after rows loop:", rows.Err())
		return user, rows.Err()
	}
	return user, nil
}

// ORG WORK
func GetProjects(c *pgx.Conn, userId int) ([]Org, error) {
	query := `SELECT o.id, o.name, m.access_tier FROM organizations o JOIN memberships m ON o.id = m.organization_id WHERE m.user_id = $1;`
	var orgs []Org
	rows, err := c.Query(context.Background(), query, userId)
	if err != nil {
		return nil, fmt.Errorf("initial query failed: %v ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var org Org
		err := rows.Scan(&org.Id, &org.Name)
		if err != nil {
			return nil, fmt.Errorf("error scanning row %v", err)
		}
		orgs = append(orgs, org)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return orgs, nil
}

func CreateProject(c *pgx.Conn, name string, ownerId string) (Org, error) {
	orgQuery := `INSERT INTO organizations (name) VALUES ($1) RETURNING id`
	memberQuery := `INSERT INTO memberships (user_id, organization_id, access_tier) VALUES ($1, $2, $3)`
	var org Org
	err := c.QueryRow(context.Background(), orgQuery, name).Scan(&org.Id)
	if err != nil {
		return org, err
	}
	_, err = c.Exec(context.Background(), memberQuery, ownerId, org.Id, "owner")
	if err != nil {
		return org, err
	}
	return org, nil
}
