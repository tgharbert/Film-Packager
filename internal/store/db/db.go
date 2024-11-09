package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id int
	Name string
	Email string
	Role string // I'm not really using this at the moment. only writing "readonly"
	Password string
	InviteStatus string
}

type ProjectUser struct {
	Id int
	Name string
	Email string
	Roles []string
	Password string
	InviteStatus string
}

type Org struct {
	Id int
	Name string
	Roles []string
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
	Members []ProjectUser
	Invited []ProjectUser
	FoundUsers []User // users on search in sidebar are placed here
}

type SelectProject struct {
	Memberships []Org
	Pending []Org
}

// written to order the roles based on preference -- optimize later?
// all of these are no eval to true????
func OrderRoles(rolesStr string) []string {
	var orderedRoles []string
	if strings.Contains(rolesStr, "owner"){
		orderedRoles = append(orderedRoles, "owner")
	}
	if strings.Contains(rolesStr, "director"){
		orderedRoles = append(orderedRoles, "director")
	}
	if strings.Contains(rolesStr, "producer"){
		orderedRoles = append(orderedRoles, "producer")
	}
	if strings.Contains(rolesStr, "writer"){
		orderedRoles = append(orderedRoles, "writer")
	}
	if strings.Contains(rolesStr, "cinematographer"){
		orderedRoles = append(orderedRoles, "cinematographer")
	}
	if strings.Contains(rolesStr, "production designer"){
		orderedRoles = append(orderedRoles, "production designer")
	}
	if strings.Contains(rolesStr, "reader"){
		orderedRoles = append(orderedRoles, "reader")
	}
	return orderedRoles
}

func CheckPasswordHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// globally available db pool to connect
var DBPool *pgxpool.Pool

func PoolConnect() {
	err := godotenv.Load()
	if err != nil {
			fmt.Println("Error loading .env file")
			panic(err)
	}
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

func CreateUser(pool *pgxpool.Pool, name string, email string, password string) (User, error) {
	var user User
	query := `INSERT INTO users (name, email, password, role) VALUES ($1, $2, $3, $4)`
	role := "readonly"
	_, err := pool.Exec(context.Background(), query, name, email, password, role)
	if err != nil {
		return user, fmt.Errorf("failed to insert user into users table: %v", err)
	}
	getNewUserQuery := `SELECT id, email, name, role FROM users WHERE email = $1`
	rows, err := pool.Query(context.Background(), getNewUserQuery, email)
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

func GetUser(pool *pgxpool.Pool, email string, password string) (User, error) {
	query := `SELECT id, email, name, role, password FROM users where email = $1`
	var user User
	rows, err := pool.Query(context.Background(), query, email)
	if err != nil {
		return user, err
	}
	defer rows.Close()
	if !rows.Next() {
		return user, fmt.Errorf("no user found with email: %s", email)
	}
	// var storedPassword string
	err = rows.Scan(&user.Id, &user.Email, &user.Name, &user.Role, &user.Password)
	if err != nil {
		return user, fmt.Errorf("error scanning row: %v", err)
	}
	err = CheckPasswordHash(user.Password, password)
	if err != nil {
		user.Password = ""
		return user, fmt.Errorf("invalid password")
	}
	if rows.Err() != nil {
		fmt.Println("Error after rows loop:", rows.Err())
		return user, rows.Err()
	}
	return user, nil
}

func GetProjects(pool *pgxpool.Pool, userId int) (SelectProject, error) {
	query := `SELECT
    o.id AS organization_id,
    o.name AS organization_name,
    array_agg(m.access_tier) AS roles,  -- Aggregate roles into an array
    m.invite_status
FROM
    organizations o
JOIN
    memberships m ON o.id = m.organization_id
WHERE
    m.user_id = $1
GROUP BY
    o.id, o.name, m.invite_status;`
	var selectProject SelectProject
	rows, err := pool.Query(context.Background(), query, userId)
	if err != nil {
		return selectProject, fmt.Errorf("initial query failed: %v ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var org Org
		err := rows.Scan(&org.Id, &org.Name, &org.Roles, &org.InviteStatus)
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

// MODIFY THIS FUNCTION TO CHECK IF THERE ARE OTHER PROJECTS WITH THIS NAME? OR NAME AND OWNER?
func CreateProject(pool *pgxpool.Pool, name string, ownerId int) (error) {
	orgQuery := `INSERT INTO organizations (name) VALUES ($1) RETURNING id, name`
	var org Org
	err := pool.QueryRow(context.Background(), orgQuery, name).Scan(&org.Id, &org.Name)
	if err != nil {
		return fmt.Errorf("failed to insert into organizations: %v", err)
	}
	memberQuery := `INSERT INTO memberships (user_id, organization_id, access_tier, invite_status) VALUES ($1, $2, $3, $4)`
	_, err = pool.Exec(context.Background(), memberQuery, ownerId, org.Id, "owner", "accepted")
	if err != nil {
		return fmt.Errorf("failed to insert into memberships: %v", err)
	}
	return nil
}

func GetProjectPageData(pool *pgxpool.Pool, projectId int) (ProjectPageData, error) {
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
),
user_roles AS (
    SELECT
        m.user_id,
        m.organization_id,
        array_agg(m.access_tier) AS roles,
        m.invite_status
    FROM memberships m
    WHERE m.organization_id = $1
    GROUP BY m.user_id, m.organization_id, m.invite_status
)
SELECT
    o.id AS project_id,
    o.name AS project_name,
    d.address AS doc_address,
    d.user_id AS doc_author,
    d.name AS doc_name,
    d.date AS doc_date,
    d.color AS doc_color,
    u.id AS user_id,
    u.name AS user_name,
    u.email AS user_email,
    ur.roles AS user_roles,  -- Array of roles
    ur.invite_status AS user_invite_status,
    d.doc_type
FROM organizations o
LEFT JOIN user_roles ur ON o.id = ur.organization_id
LEFT JOIN users u ON ur.user_id = u.id
LEFT JOIN doc_types d ON o.id = d.organization_id
WHERE o.id = $1  -- Ensure we only fetch data for the given organization
ORDER BY o.id;
`
	rows, err := pool.Query(context.Background(), query, projectId)
	if err != nil {
		return ProjectPageData{}, err
	}
	defer rows.Close()
	var projectData ProjectPageData
	projectMap := make(map[string]DocInfo)

	for rows.Next() {
		var docName, userName, userEmail, docType sql.NullString
		var docAddress, docColor sql.NullString
		var userId, docAuthor sql.NullInt32
		var userRoles sql.NullString
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
			&userRoles,
			&inviteStatus,
			&docType,
		)
		if err != nil {
			return projectData, fmt.Errorf("error scanning row: %w", err)
		}
		roles := []string{}
		if userRoles.Valid {
			rolesStr := userRoles.String
			rolesStr = strings.Trim(rolesStr, "{}")
			if rolesStr != "" {
				roles = strings.Split(rolesStr, ",")
			}
		}
		roles = OrderRoles(roles[0])
		if inviteStatus.String == "pending" {
			projectData.Invited = append(projectData.Invited, ProjectUser{
				Id:    int(userId.Int32),  // Convert sql.NullInt32 to int
				Name:  userName.String,
				Email: userEmail.String,
				Roles: roles,
				InviteStatus: inviteStatus.String,
			})
		}

		if inviteStatus.String == "accepted" {
			projectData.Members = append(projectData.Members, ProjectUser{
				Id:    int(userId.Int32),  // Convert sql.NullInt32 to int
				Name:  userName.String,
				Email: userEmail.String,
				Roles: roles,
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

func SearchForUsers(pool *pgxpool.Pool, queryString string) ([]User, error) {
	query := `SELECT id, name FROM users WHERE name ILIKE '%' || $1 || '%'`
	rows, err := pool.Query(context.Background(), query, queryString)
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

func InviteUserToOrg(pool *pgxpool.Pool, memberId int, organizationId int, role string) ([]User, error) {
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
	rows, err := pool.Query(context.Background(), query, memberId, organizationId, role, "pending")
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

func JoinOrg(pool *pgxpool.Pool, projectId int, memberId int, role string) (error) {
	updateQuery := `
		UPDATE memberships
		SET invite_status = 'accepted'
		WHERE organization_id = $1 AND user_id = $2 and access_tier = $3;
	`
	_, err := pool.Exec(context.Background(), updateQuery, projectId, memberId, role)
	if err != nil {
		return fmt.Errorf("error updating membership status: %v", err)
	}
	return nil
}

func DeleteOrg(pool *pgxpool.Pool, orgId int, userId int) (error) {
	deleteProjectQuery := `DELETE FROM organizations WHERE id = $1;`
	_, err := pool.Query(context.Background(), deleteProjectQuery, orgId)
	if err != nil {
		return fmt.Errorf("failed to delete project: %v", err)
	}
	return nil
}