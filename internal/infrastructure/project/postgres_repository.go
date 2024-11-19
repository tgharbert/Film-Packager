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
            array_agg(m.role) AS roles,
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
		var project = &domain.ProjectOverview{}
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
