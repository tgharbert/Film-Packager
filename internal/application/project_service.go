package application

import (
	"context"
	"filmPackager/internal/domain"
	"fmt"
)

type ProjectRepository interface {
	GetProjectsForUserSelection(ctx context.Context, userId int) ([]*domain.ProjectOverview, error)
	// GetProjectById(projectId int) (*domain.Project, error)
}

type ProjectService struct {
	projRepo ProjectRepository
}

func NewProjectService(projRepo ProjectRepository) *ProjectService {
	return &ProjectService{projRepo: projRepo}
}

// should this take in the User then get the projects and sort them for the user??
func (s *ProjectService) GetUsersProjects(ctx context.Context, user *domain.User) (*domain.User, error) {
	// do auth work here?
	projects, err := s.projRepo.GetProjectsForUserSelection(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting projects from db: %v", err)
	}
	for _, project := range projects {
		// sort the roles in each project here as well
		if project.Status == "invited" {
			user.Invited = append(user.Invited, *project)
		}
		if project.Status == "member" {
			user.Memberships = append(user.Invited, *project)
		}
	}
	return user, nil
}
