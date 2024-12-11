package infrastructure

import (
	"context"
	"fmt"

	"filmPackager/internal/domain/membership"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresMembershipRepository struct {
	db *pgxpool.Pool
}

func NewPostgresMembershipRepository(db *pgxpool.Pool) *PostgresMembershipRepository {
	return &PostgresMembershipRepository{db: db}
}

func (r *PostgresMembershipRepository) GetProjectMemberships(ctx context.Context, projectId uuid.UUID) ([]membership.Membership, error) {
	query := `
		SELECT
	id,
	user_id,
	organization_id,
	access_tier,
	invite_status 
	FROM
	memberships
	WHERE organization_id = $1`
	var memberships []membership.Membership
	rows, err := r.db.Query(ctx, query, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project memberships: %v", err)
	}
	for rows.Next() {
		m := membership.Membership{}
		err := rows.Scan(&m.ID, &m.UserID, &m.ProjectID, &m.Roles, &m.InviteStatus)
		if err != nil {
			return nil, fmt.Errorf("error scanning project membership row: %v", err)
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}

func (r *PostgresMembershipRepository) CreateMembership(ctx context.Context, m *membership.Membership) error {
	query := `
		INSERT INTO memberships (id, user_id, organization_id, access_tier, invite_status) 
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, m.ID, m.UserID, m.ProjectID, m.Roles, m.InviteStatus)
	if err != nil {
		return fmt.Errorf("error creating membership: %v", err)
	}
	return nil
}

func (r *PostgresMembershipRepository) DeleteMembership(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error {
	query := `
		DELETE FROM memberships 
		WHERE user_id = $1 AND project_id = $2`
	_, err := r.db.Exec(ctx, query, userID, projectID)
	if err != nil {
		return fmt.Errorf("error deleting membership: %v", err)
	}
	return nil
}

func (r *PostgresMembershipRepository) GetMembership(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) (*membership.Membership, error) {
	query := `
		SELECT
	user_id,
	project_id,
	user_name,
	user_email,
	roles,
	invite_status 
	FROM
	memberships
	WHERE project_id = $1 AND user_id = $2`
	var m membership.Membership
	err := r.db.QueryRow(ctx, query, projectId, userId).Scan(&m.UserID, &m.ProjectID, &m.UserName, &m.UserEmail, &m.Roles, &m.InviteStatus)
	if err != nil {
		return nil, fmt.Errorf("error getting membership: %v", err)
	}
	return &m, nil
}

func (r *PostgresMembershipRepository) GetUserMembershipsForProject(ctx context.Context, userId uuid.UUID, projectId uuid.UUID) (*membership.Membership, error) {
	query := `
		SELECT
	user_id,
	project_id,
	user_name,
	user_email,
	roles,
	invite_status 
	FROM
	memberships
	WHERE project_id = $1 AND user_id = $2`
	var m membership.Membership
	err := r.db.QueryRow(ctx, query, projectId, userId).Scan(&m.UserID, &m.ProjectID, &m.UserName, &m.UserEmail, &m.Roles, &m.InviteStatus)
	if err != nil {
		return nil, fmt.Errorf("error getting membership: %v", err)
	}
	return &m, nil
}

func (r *PostgresMembershipRepository) GetProjectIDsForUser(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT
	project_id 
	FROM
	memberships
	WHERE user_id = $1`
	var projectIds []uuid.UUID
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting project ids for user: %v", err)
	}
	for rows.Next() {
		var projectId uuid.UUID
		err := rows.Scan(&projectId)
		if err != nil {
			return nil, fmt.Errorf("error scanning project id row: %v", err)
		}
		projectIds = append(projectIds, projectId)
	}
	return projectIds, nil
}

func (r *PostgresMembershipRepository) GetAllUserMemberships(ctx context.Context, userId uuid.UUID) ([]membership.Membership, error) {
	query := `
		SELECT
	id,
	user_id,
	organization_id,
	access_tier,
	invite_status 
	FROM
	memberships
	WHERE user_id = $1`
	var memberships []membership.Membership
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting all user memberships: %v", err)
	}
	for rows.Next() {
		m := membership.Membership{}
		err := rows.Scan(&m.ID, &m.UserID, &m.ProjectID, &m.Roles, &m.InviteStatus)
		if err != nil {
			return nil, fmt.Errorf("error scanning user membership row: %v", err)
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}
