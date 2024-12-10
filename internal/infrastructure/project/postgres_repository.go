package infrastructure

import (
	"context"
	"errors"

	"fmt"

	"filmPackager/internal/domain/project"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProjectRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProjectRepository(db *pgxpool.Pool) *PostgresProjectRepository {
	return &PostgresProjectRepository{db: db}
}

func (r *PostgresProjectRepository) GetProjectsForUserSelection(ctx context.Context, userId uuid.UUID) ([]project.ProjectOverview, error) {
	query := `
        SELECT
    o.id,
    o.name,
    m.access_tier AS roles, -- Directly retrieve the array from the column
    m.invite_status
FROM
    organizations o
JOIN
    memberships m ON o.id = m.organization_id
WHERE
    m.user_id = $1;
    `
	var projects []project.ProjectOverview
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		fmt.Println("error in query: ", err)
		return nil, fmt.Errorf("error querying db: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var project project.ProjectOverview
		var roles []string
		err = rows.Scan(&project.ID, &project.Name, &roles, &project.Status)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %v", err)
		}
		project.Roles = roles
		projects = append(projects, project)
	}
	return projects, nil
}

func (r *PostgresProjectRepository) CreateNewProject(ctx context.Context, p *project.Project, ownerId uuid.UUID) (*project.ProjectOverview, error) {
	// ensure that both transactions fail or succeed together
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()
	orgQuery := `INSERT INTO organizations (id, owner_id, created_at, updated_at, name) VALUES ($1, $2, $3, $4, $5) RETURNING id, name`
	var project project.ProjectOverview
	err = r.db.QueryRow(ctx, orgQuery, p.ID, p.OwnerID, p.CreatedAt, p.LastUpdateAt, p.Name).Scan(&project.ID, &project.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to insert into organizations: %v", err)
	}
	accessTiers := []string{"owner"}
	memberQuery := `INSERT INTO memberships (user_id, organization_id, access_tier, invite_status) VALUES ($1, $2, $3, $4) RETURNING access_tier, invite_status`
	_, err = r.db.Exec(ctx, memberQuery, ownerId, project.ID, accessTiers, "accepted")
	if err != nil {
		return nil, fmt.Errorf("failed to insert into memberships: %v", err)
	}
	project.Roles = accessTiers
	return &project, nil
}

func (r *PostgresProjectRepository) DeleteProject(ctx context.Context, projectId uuid.UUID) error {
	deleteProjectQuery := `DELETE FROM organizations WHERE id = $1;`
	_, err := r.db.Exec(ctx, deleteProjectQuery, projectId)
	if err != nil {
		return fmt.Errorf("failed to delete project and fetch remaining projects: %v", err)
	}
	return nil
}

func (r *PostgresProjectRepository) GetProjectDetails(ctx context.Context, projectId uuid.UUID) (*project.Project, error) {
	var project project.Project
	query := `SELECT id, name FROM organizations WHERE id = $1`
	err := r.db.QueryRow(ctx, query, projectId).Scan(&project.ID, &project.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting project from db: %v", err)
	}
	return &project, nil
}

func (r *PostgresProjectRepository) GetProjectUsers(ctx context.Context, projectId uuid.UUID) ([]project.ProjectMembership, error) {
	query := `SELECT
    u.id AS user_id,
    u.name AS user_name,
    u.email AS user_email,
		m.invite_status AS status,
    m.access_tier AS user_roles
FROM
    memberships m
JOIN
    users u ON m.user_id = u.id
WHERE
    m.organization_id = $1
`
	var members []project.ProjectMembership
	rows, err := r.db.Query(ctx, query, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project's users: %v", err)
	}
	for rows.Next() {
		member := &project.ProjectMembership{}
		err := rows.Scan(&member.UserID, &member.UserName, &member.UserEmail, &member.InviteStatus, &member.Roles)
		if err != nil {
			return nil, err
		}
		members = append(members, *member)
	}
	return members, nil
}

func (r *PostgresProjectRepository) SearchForUsers(ctx context.Context, name string) ([]project.ProjectMembership, error) {
	query := `SELECT id, name FROM users WHERE name ILIKE '%' || $1 || '%'`
	rows, err := r.db.Query(context.Background(), query, name)
	var users []project.ProjectMembership
	if err != nil {
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		user := &project.ProjectMembership{}
		err := rows.Scan(&user.UserID, &user.UserName)
		if err != nil {
			return nil, fmt.Errorf("error scanning user row %v", err)
		}
		users = append(users, *user)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return users, nil
}

func (r *PostgresProjectRepository) InviteMember(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error {
	query := `INSERT INTO memberships (user_id, organization_id) VALUES ($1, $2)`
	_, err := r.db.Exec(ctx, query, userId, projectId)
	if err != nil {
		return fmt.Errorf("error inviting user to project: %v", err)
	}
	return nil
}

func (r *PostgresProjectRepository) JoinProject(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error {
	query := `UPDATE memberships SET invite_status = 'accepted' WHERE user_id = $1 AND organization_id = $2`
	_, err := r.db.Exec(ctx, query, userId, projectId)
	if err != nil {
		return fmt.Errorf("error joining project: %v", err)
	}
	return nil
}

func (r *PostgresProjectRepository) GetProjectUser(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) (*project.ProjectMembership, error) {
	query := `SELECT
    u.id AS user_id,
    u.name AS user_name,
    u.email AS user_email,
    m.invite_status AS status,
    m.access_tier AS user_role
FROM
    memberships m
JOIN
    users u ON m.user_id = u.id
WHERE
    m.organization_id = $1 AND m.user_id = $2;
`
	user := project.ProjectMembership{}
	err := r.db.QueryRow(ctx, query, projectId, userId).Scan(&user.UserID, &user.UserName, &user.UserEmail, &user.InviteStatus, &user.Roles)
	// check if there are no rows for this user/project
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, project.ErrMemberNotFound
		}
		return nil, fmt.Errorf("error getting user from project: %v", err)
	}
	return &user, nil
}

func (r *PostgresProjectRepository) UpdateMemberRoles(ctx context.Context, projectId uuid.UUID, userId uuid.UUID, role string) error {
	query := `UPDATE memberships SET access_tier = array_append(array_remove(access_tier, $1), $2) WHERE organization_id = $3 AND user_id = $4`
	_, err := r.db.Query(ctx, query, "reader", role, projectId, userId)
	if err != nil {
		return fmt.Errorf("error updating member roles: %v", err)
	}
	return nil
}
