package infrastructure

import (
	"context"

	"fmt"

	"filmPackager/internal/domain/project"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProjectRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProjectRepository(db *pgxpool.Pool) *PostgresProjectRepository {
	return &PostgresProjectRepository{db: db}
}

func (r *PostgresProjectRepository) GetProjectsForUserSelection(ctx context.Context, userId int) ([]*project.ProjectOverview, error) {
	query := `
        SELECT
            o.id,
            o.name,
            array_agg(m.access_tier) AS roles,
            m.invite_status
        FROM
            organizations o
        JOIN
            memberships m ON o.id = m.organization_id
        WHERE
            m.user_id = $1
        GROUP BY
            o.id, o.name, m.invite_status;
    `
	var projects []*project.ProjectOverview
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("error querying db: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var project project.ProjectOverview
		var roles []string
		err = rows.Scan(&project.Id, &project.Name, &roles, &project.Status)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %v", err)
		}
		project.Roles = roles
		projects = append(projects, &project)
	}
	return projects, nil
}

func (r *PostgresProjectRepository) CreateNewProject(ctx context.Context, projectName string, ownerId int) (*project.ProjectOverview, error) {
	orgQuery := `INSERT INTO organizations (name) VALUES ($1) RETURNING id, name`
	var project project.ProjectOverview
	err := r.db.QueryRow(context.Background(), orgQuery, projectName).Scan(&project.Id, &project.Name)
	project.Roles = append(project.Roles, "owner")
	// project.Status = "accepted"
	if err != nil {
		return nil, fmt.Errorf("failed to insert into organizations: %v", err)
	}
	memberQuery := `INSERT INTO memberships (user_id, organization_id, access_tier, invite_status) VALUES ($1, $2, $3, $4) RETURNING access_tier, invite_status`
	_, err = r.db.Exec(context.Background(), memberQuery, ownerId, project.Id, "owner", "accepted")
	if err != nil {
		return nil, fmt.Errorf("failed to insert into memberships: %v", err)
	}
	return &project, nil
}

func (r *PostgresProjectRepository) DeleteProject(ctx context.Context, projectId int) error {
	deleteProjectQuery := `DELETE FROM organizations WHERE id = $1;`
	_, err := r.db.Exec(ctx, deleteProjectQuery, projectId)
	if err != nil {
		return fmt.Errorf("failed to delete project and fetch remaining projects: %v", err)
	}
	return nil
}

func (r *PostgresProjectRepository) GetProjectDetails(ctx context.Context, projectId int) (*project.Project, error) {
	var project project.Project
	query := `SELECT id, name FROM organizations WHERE id = $1`
	err := r.db.QueryRow(ctx, query, projectId).Scan(&project.Id, &project.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting project from db: %v", err)
	}
	return &project, nil
}

func (r *PostgresProjectRepository) GetProjectUsers(ctx context.Context, projectId int) ([]*project.ProjectMembership, error) {
	query := `SELECT
    u.id AS user_id,
    u.name AS user_name,
    u.email AS user_email,
		m.invite_status AS status,
    array_agg(m.access_tier) AS user_roles
FROM
    memberships m
JOIN
    users u ON m.user_id = u.id
WHERE
    m.organization_id = $1
GROUP BY
    u.id, u.name, u.email, m.invite_status;
`
	var members []*project.ProjectMembership
	rows, err := r.db.Query(ctx, query, projectId)
	if err != nil {
		fmt.Println("here is the error: ", err)

		return nil, fmt.Errorf("error getting project's users: %v", err)
	}
	for rows.Next() {
		member := &project.ProjectMembership{}
		err := rows.Scan(&member.UserId, &member.UserName, &member.UserEmail, &member.InviteStatus, &member.Roles)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

func (r *PostgresProjectRepository) SearchForUsers(ctx context.Context, name string) ([]*project.ProjectMembership, error) {
	query := `SELECT id, name FROM users WHERE name ILIKE '%' || $1 || '%'`
	rows, err := r.db.Query(context.Background(), query, name)
	var users []*project.ProjectMembership
	if err != nil {
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		user := &project.ProjectMembership{}
		err := rows.Scan(&user.UserId, &user.UserName)
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

func (r *PostgresProjectRepository) InviteMember(ctx context.Context, userId int, projectId int, role string) error {
	// query to write membership to db
	query := `INSERT INTO memberships (user_id, organization_id, access_tier) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, userId, projectId, role)
	if err != nil {
		return fmt.Errorf("error inviting user to project: %v", err)
	}
	return nil
}

func (r *PostgresProjectRepository) JoinProject(ctx context.Context, projectId int, userId int, role string) error {
	query := `UPDATE memberships SET invite_status = 'accepted' WHERE user_id = $1 AND organization_id = $2 AND access_tier = $3`
	fmt.Printf("Debug: userId=%d, projectId=%d, role=%s\n", userId, projectId, role)
	_, err := r.db.Exec(ctx, query, userId, projectId, role)
	if err != nil {
		return fmt.Errorf("error joining project: %v", err)
	}
	return nil
}
