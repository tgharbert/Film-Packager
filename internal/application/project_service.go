package application

import (
	"context"
	"filmPackager/internal/domain"
	"fmt"
)

type ProjectRepository interface {
	GetProjectsForUserSelection(ctx context.Context, userId int) ([]*domain.ProjectOverview, error)
	// GetProjectById(ctx, projectId int) (*domain.Project, error)
	CreateNewProject(ctx context.Context, projectName string, userId int) (*domain.ProjectOverview, error)
	DeleteProject(ctx context.Context, projectId int) error
	GetProjectDetails(ctx context.Context, projectId int) (*domain.Project, error)
	GetProjectUsers(ctx context.Context, projectId int) ([]*domain.ProjectMembership, error)
	SearchForUsers(ctx context.Context, userName string) ([]*domain.ProjectMembership, error)
}

type ProjectService struct {
	projRepo ProjectRepository
}

func NewProjectService(projRepo ProjectRepository) *ProjectService {
	return &ProjectService{
		projRepo: projRepo,
	}
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
		if project.Status == "accepted" {
			user.Memberships = append(user.Memberships, *project)
		}
	}
	return user, nil
}

func (s *ProjectService) CreateNewProject(ctx context.Context, projectName string, userId int) (*domain.ProjectOverview, error) {
	project, err := s.projRepo.CreateNewProject(ctx, projectName, userId)
	if err != nil {
		return nil, fmt.Errorf("error with project creation: %v", err)
	}
	return project, nil
}

// should this be in the user service??
func (s *ProjectService) DeleteProject(ctx context.Context, projectId int, user *domain.User) (*domain.User, error) {
	err := s.projRepo.DeleteProject(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error deleting projects from db: %v", err)
	}
	projects, err := s.projRepo.GetProjectsForUserSelection(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user projects: %v", err)
	}
	// var user *domain.User
	for _, project := range projects {
		// sort the roles in each project here as well
		if project.Status == "invited" {
			user.Invited = append(user.Invited, *project)
		}
		if project.Status == "accepted" {
			user.Memberships = append(user.Memberships, *project)
		}
	}
	return user, nil
}

func (s *ProjectService) GetProjectDetails(ctx context.Context, projectId int) (*domain.Project, error) {
	// get the project from the db
	project, err := s.projRepo.GetProjectDetails(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project from db: %v", err)
	}
	members, err := s.projRepo.GetProjectUsers(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project users from db: %v", err)
	}
	// i have the data, but I need to sort the roles
	// I would also like to sort the members if possible
	for _, member := range members {
		project.Members = append(project.Members, *member)
		// fmt.Println("members: ", *member)
	}
	// sort the roles??
	return project, nil
}

func (s *ProjectService) SearchForUsers(ctx context.Context, userName string) ([]*domain.ProjectMembership, error) {
	users, err := s.projRepo.SearchForUsers(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("error searching for users: %v", err)
	}
	return users, nil
}
