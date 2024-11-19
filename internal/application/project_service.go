package application

import (
	"context"
	"filmPackager/internal/domain"
)

type ProjectRepository interface {
	GetProjectsForUserSelection(ctx context.Context, userId int) ([]*domain.ProjectOverview, error)
	GetProjectById(projectId int) (*domain.Project, error)
}

type ProjectService struct {
	projRepo ProjectRepository
}

func NewProjectRepository(projRepo *ProjectRepository) *ProjectService {
	return &ProjectService{projRepo: *projRepo}
}

func (s *ProjectService) GetProjectsForUserSelection(userId int) ([]*domain.ProjectOverview, error) {
	var projects []*domain.ProjectOverview
	// query the database for all of the projects

	// sort the roles for each project

	// sort the projects as desired - invited, member

	return projects, nil
}
