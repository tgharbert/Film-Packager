package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

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
	InviteStatus string
}

type Org struct {
	Id int
	Name string
	Role string
	InviteStatus string
}

type Project struct {
	Id int
	Name string
	Script DocInfo
	Logline DocInfo
	Synopsis DocInfo
	PitchDeck DocInfo
	Schedule DocInfo
	Budget DocInfo
	DirectorStatement DocInfo
	Shotlist DocInfo
	Lookbook DocInfo
	Bios DocInfo
}

type DocInfo struct {
	Id int
	Name string
	Date time.Time
	Author int // user id of author??
	Color string
}

type ProjectPageData struct {
	Project Project
	Members []User
	FoundUsers []User // users on search in sidebar are placed here
}

type SelectProject struct {
	Memberships []Org
	Pending []Org
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
	getNewUserQuery := `SELECT id, email, name, role FROM users WHERE email = $1`
	rows, err := c.Query(context.Background(), getNewUserQuery, email)
	if err != nil {
		return user, err
	}
	defer rows.Close()
	if !rows.Next() {
		return user, fmt.Errorf("no user found with email: %s", email)
	}
	err = rows.Scan(&user.Id, &user.Email, &user.Name, &user.Role)
	if err != nil {
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

// MODIFY THIS QUERY TO RETURN SELECTPROJECT STRUCT TO SEPERATE MEMBERSHIPS - INCLUDE ACCESS TIERS
func GetProjects(c *pgx.Conn, userId int) (SelectProject, error) {
	query := `SELECT o.id, o.name, m.access_tier, m.invite_status FROM organizations o JOIN memberships m ON o.id = m.organization_id WHERE m.user_id = $1;`
	var selectProject SelectProject
	rows, err := c.Query(context.Background(), query, userId)
	if err != nil {
		return selectProject, fmt.Errorf("initial query failed: %v ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var org Org
		err := rows.Scan(&org.Id, &org.Name, &org.Role, &org.InviteStatus)
		if err != nil {
			return selectProject, fmt.Errorf("error scanning row %v", err)
		}
		if org.InviteStatus == "pending" {
			selectProject.Pending = append(selectProject.Pending, org)
		} else if org.InviteStatus == "accepted" {
			selectProject.Memberships = append(selectProject.Memberships, org)
		}
	}
	if rows.Err() != nil {
		return selectProject, rows.Err()
	}
	return selectProject, nil
}

func CreateProject(c *pgx.Conn, name string, ownerId int) (Org, error) {
	// Begin transaction
	tx, err := c.Begin(context.Background())
	if err != nil {
		return Org{}, fmt.Errorf("could not begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background()) // Rollback if any step fails
	// Insert organization
	orgQuery := `INSERT INTO organizations (name) VALUES ($1) RETURNING id, name`
	var org Org
	err = tx.QueryRow(context.Background(), orgQuery, name).Scan(&org.Id, &org.Name)
	if err != nil {
		return org, fmt.Errorf("failed to insert into organizations: %v", err)
	}
	// Insert into memberships with the role "owner"
	memberQuery := `INSERT INTO memberships (user_id, organization_id, access_tier, invite_status) VALUES ($1, $2, $3, $4) RETURNING access_tier`
	err = tx.QueryRow(context.Background(), memberQuery, ownerId, org.Id, "owner", "accepted").Scan(&org.Role)
	if err != nil {
		return org, fmt.Errorf("failed to insert into memberships: %v", err)
	}
	// Commit transaction
	err = tx.Commit(context.Background())
	if err != nil {
		return org, fmt.Errorf("failed to commit transaction: %v", err)
	}
	return org, nil
}

// THIS IS WHERE I'M AT
func GetProjectPageData(c *pgx.Conn, projectId int) (ProjectPageData, error) {
	// I want to return all users attached to a project, all docs attached to the project -- not really too difficult
	query := `WITH doc_types AS (
    SELECT
        organization_id,
        user_id,
        name,
        date,
        color,
        address,
        CASE
            WHEN name = 'Script' THEN 'Script'
            WHEN name = 'Logline' THEN 'Logline'
            WHEN name = 'Synopsis' THEN 'Synopsis'
            WHEN name = 'Pitch Deck' THEN 'Pitch Deck'
            WHEN name = 'Schedule' THEN 'Schedule'
            WHEN name = 'Budget' THEN 'Budget'
            WHEN name = 'Director Statement' THEN 'DirectorStatement'
            WHEN name = 'Shotlist' THEN 'Shotlist'
            WHEN name = 'Lookbook' THEN 'Lookbook'
            WHEN name = 'Bios' THEN 'Bios'
            ELSE NULL
        END AS doc_type
    FROM documents
)
SELECT
    o.id AS project_id,
    o.name AS project_name,
    d.address AS doc_address,
    d.user_id AS doc_author,
    d.name AS doc_name,
    d.date AS doc_date,
    d.color AS doc_color,
    m.user_id,
    u.name AS user_name,
    u.email AS user_email,
		m.access_tier AS user_role,
		m.invite_status AS user_invite_status,
		d.doc_type
FROM organizations o
LEFT JOIN memberships m ON o.id = m.organization_id
LEFT JOIN users u ON m.user_id = u.id
LEFT JOIN doc_types d ON o.id = d.organization_id
WHERE o.id = $1 -- assuming you're passing the organization ID as a parameter
ORDER BY o.id;
	`
	rows, err := c.Query(context.Background(), query, projectId)
	if err != nil {
		fmt.Println("Here is the error!", err)
		return ProjectPageData{}, err
	}
	defer rows.Close()
	var projectData ProjectPageData
	projectMap := make(map[string]DocInfo)

	for rows.Next() {
		var docName, userName, userEmail, docType sql.NullString
		var docAddress, docColor sql.NullString
		var userId, docAuthor sql.NullInt32
		var userRole sql.NullString
		var inviteStatus sql.NullString
		// projectId int
		var docDate sql.NullTime

		err := rows.Scan(
			&projectData.Project.Id,
			&projectData.Project.Name,
			&docAddress,
			&docAuthor,
			&docName,
			&docDate,
			&docColor,
			&userId,
			&userName,
			&userEmail,
			&userRole,
			&inviteStatus,
			&docType,
		)
		if err != nil {
			return projectData, err
		}

		if userName.Valid && userEmail.Valid {
			// Add members
			projectData.Members = append(projectData.Members, User{
				Id:    int(userId.Int32),  // Convert sql.NullInt32 to int
				Name:  userName.String,
				Email: userEmail.String,
				Role: userRole.String,
				InviteStatus: inviteStatus.String,
			})
		}

		if docType.Valid && docName.Valid && docDate.Valid {
			// Map documents to the project by their docType
			projectMap[docType.String] = DocInfo{
				Id:     int(docAuthor.Int32),  // Convert sql.NullInt32 to int
				Name:   docName.String,
				Date:   docDate.Time,
				Author: int(docAuthor.Int32),  // Convert sql.NullInt32 to int
				Color:  docColor.String,
			}
		}
	}

	// Assign each document type to the project struct
	projectData.Project.Script = projectMap["Script"]
	projectData.Project.Logline = projectMap["Logline"]
	projectData.Project.Synopsis = projectMap["Synopsis"]
	projectData.Project.PitchDeck = projectMap["Pitch Deck"]
	projectData.Project.Schedule = projectMap["Schedule"]
	projectData.Project.Budget = projectMap["Budget"]
	projectData.Project.DirectorStatement = projectMap["DirectorStatement"]
	projectData.Project.Shotlist = projectMap["Shotlist"]
	projectData.Project.Lookbook = projectMap["Lookbook"]
	projectData.Project.Bios = projectMap["Bios"]

	if rows.Err() != nil {
		return projectData, rows.Err()
	}
	return projectData, nil
}

func SearchForUsers(c *pgx.Conn, queryString string) ([]User, error) {
	query := `SELECT id, name FROM users WHERE name ILIKE '%' || $1 || '%'`
	rows, err := c.Query(context.Background(), query, queryString)
	var users []User
	if err != nil {
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Name)
		if err != nil {
			return nil, fmt.Errorf("error scanning user row %v", err)
		}
		users = append(users, user)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return users, nil
}

func InviteUserToOrg(c *pgx.Conn, memberId int, organizationId int, role string) ([]User, error) {
	query := `
	WITH new_membership AS (
			INSERT INTO memberships (user_id, organization_id, access_tier, invite_status)
			VALUES ($1, $2, $3, $4)
			RETURNING user_id, organization_id, access_tier, invite_status
	)
	SELECT
			new_membership.user_id,
			new_membership.access_tier AS role,
			new_membership.invite_status,
			users.name,
			users.email
	FROM
			new_membership
	JOIN
			users ON users.id = new_membership.user_id
	UNION ALL
	SELECT
			memberships.user_id,
			memberships.access_tier,
			memberships.invite_status,
			users.name,
			users.email
	FROM
			memberships
	JOIN
			users ON users.id = memberships.user_id
	WHERE
			memberships.organization_id = $2;`
	var users []User
	rows, err := c.Query(context.Background(), query, memberId, organizationId, role, "pending")
	if err != nil {
		return users, fmt.Errorf("error querying: %v", err)
	}
	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Role, &user.InviteStatus, &user.Name, &user.Email)
		if err != nil {
			return users, fmt.Errorf("error scanning row %v", err)
		}
		users = append(users, user)
	}
	if rows.Err() != nil {
		return users, rows.Err()
	}
	return users, nil
}

func JoinOrg(c *pgx.Conn, projectId int, memberId int, role string) (SelectProject, error) {
	var projects SelectProject
	updateQuery := `
		UPDATE memberships
		SET invite_status = 'accepted'
		WHERE organization_id = $1 AND user_id = $2 and access_tier = $3;
	`
	_, err := c.Exec(context.Background(), updateQuery, projectId, memberId, role)
	if err != nil {
		return projects, fmt.Errorf("error updating membership status: %v", err)
	}
	// Step 2: Retrieve all projects for the user
	selectQuery := `
		SELECT
			o.id AS organization_id,
			o.name AS organization_name,
			m.access_tier,
			m.invite_status
		FROM
			organizations o
		JOIN
			memberships m
		ON
			o.id = m.organization_id
		WHERE
			m.user_id = $1;
	`
	rows, err := c.Query(context.Background(), selectQuery, memberId)
	if err != nil {
		return SelectProject{}, fmt.Errorf("error querying projects: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var project Org
		err := rows.Scan(&project.Id, &project.Name, &project.Role, &project.InviteStatus)
		if err != nil {
			return projects, fmt.Errorf("error scanning row: %v", err)
		}
		if project.InviteStatus == "pending" {
			projects.Pending = append(projects.Pending, project)
		}
		if project.InviteStatus == "accepted" {
			projects.Memberships = append(projects.Memberships, project)
		}
	}
	if rows.Err() != nil {
		return projects, rows.Err()
	}
	return projects, nil
}

func DeleteOrg(c *pgx.Conn, orgId int, userId int) (SelectProject, error) {
	// query should delete specified project and all related material, then return remaining ones
	query := `SELECT
    o.id AS organization_id,
    o.name AS organization_name,
    m.access_tier,
    m.invite_status
FROM
    organizations o
JOIN
    memberships m ON o.id = m.organization_id
WHERE
    m.user_id = $1;`
	var projects SelectProject
	rows, err := c.Query(context.Background(), query, orgId)
	if err != nil {
		return SelectProject{}, fmt.Errorf("error querying projects: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var project Org
		err := rows.Scan(&project.Id, &project.Name, &project.Role, &project.InviteStatus)
		if err != nil {
			return projects, fmt.Errorf("error scanning row: %v", err)
		}
		if project.InviteStatus == "pending" {
			projects.Pending = append(projects.Pending, project)
		}
		if project.InviteStatus == "accepted" {
			projects.Memberships = append(projects.Memberships, project)
		}
	}
	if rows.Err() != nil {
		return projects, rows.Err()
	}
	return projects, nil
}