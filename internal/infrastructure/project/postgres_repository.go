package infrastructure

import (
	"context"
	"filmPackager/internal/domain"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProjectRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProjectRepository(db *pgxpool.Pool) *PostgresProjectRepository {
	return &PostgresProjectRepository{db: db}
}

func (r *PostgresProjectRepository) GetProjectsForUserSelection(ctx context.Context, userId int) ([]*domain.ProjectOverview, error) {
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
	var projects []*domain.ProjectOverview
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("error querying db: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var project domain.ProjectOverview
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

func (r *PostgresProjectRepository) CreateNewProject(ctx context.Context, projectName string, ownerId int) (*domain.ProjectOverview, error) {
	orgQuery := `INSERT INTO organizations (name) VALUES ($1) RETURNING id, name`
	var project domain.ProjectOverview
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
