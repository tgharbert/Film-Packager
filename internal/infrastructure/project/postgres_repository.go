package infrastructure

import (
	"context"

	"fmt"

	"filmPackager/internal/domain/project"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProjectRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProjectRepository(db *pgxpool.Pool) *PostgresProjectRepository {
	return &PostgresProjectRepository{db: db}
}

func (r *PostgresProjectRepository) GetProjectsByMembershipIDs(ctx context.Context, projectIds []uuid.UUID) ([]project.Project, error) {
	// should take in an array of ids and return an array of projects?
	query := `SELECT id, name FROM organizations WHERE id = ANY($1)`
	var projects []project.Project
	rows, err := r.db.Query(ctx, query, projectIds)
	if err != nil {
		return nil, fmt.Errorf("error getting projects from db: %v", err)
	}
	for rows.Next() {
		var p project.Project
		err := rows.Scan(&p.ID, &p.Name)
		if err != nil {
			return nil, fmt.Errorf("error scanning project row: %v", err)
		}
		projects = append(projects, p)
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

func (r *PostgresProjectRepository) UpdateMemberRoles(ctx context.Context, projectId uuid.UUID, userId uuid.UUID, role string) error {
	query := `UPDATE memberships SET access_tier = array_append(array_remove(access_tier, $1), $2) WHERE organization_id = $3 AND user_id = $4`
	_, err := r.db.Query(ctx, query, "reader", role, projectId, userId)
	if err != nil {
		return fmt.Errorf("error updating member roles: %v", err)
	}
	return nil
}
