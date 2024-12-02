package application

import (
	"context"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"

	"fmt"
)

type ProjectService struct {
	projRepo project.ProjectRepository
}

func NewProjectService(projRepo project.ProjectRepository) *ProjectService {
	return &ProjectService{
		projRepo: projRepo,
	}
}

// should this take in the User then get the projects and sort them for the user??
func (s *ProjectService) GetUsersProjects(ctx context.Context, user *user.User) (*user.User, error) {
	// do auth work here?
	projects, err := s.projRepo.GetProjectsForUserSelection(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting projects from db: %v", err)
	}
	for _, project := range projects {
		// sort the roles in each project here as well
		if project.Status == "pending" {
			user.Invited = append(user.Invited, *project)
		}
		if project.Status == "accepted" {
			user.Memberships = append(user.Memberships, *project)
		}
	}
	return user, nil
}

func (s *ProjectService) CreateNewProject(ctx context.Context, projectName string, userId int) (*project.ProjectOverview, error) {
	project, err := s.projRepo.CreateNewProject(ctx, projectName, userId)
	if err != nil {
		return nil, fmt.Errorf("error with project creation: %v", err)
	}
	return project, nil
}

// should this be in the user service??
func (s *ProjectService) DeleteProject(ctx context.Context, projectId int, user *user.User) (*user.User, error) {
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

func (s *ProjectService) GetProjectDetails(ctx context.Context, projectId int) (*project.Project, error) {
	// get the project from the db
	projectDetails, err := s.projRepo.GetProjectDetails(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project from db: %v", err)
	}
	members, err := s.projRepo.GetProjectUsers(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project users from db: %v", err)
	}
	for _, member := range members {
		member.Roles = project.SortRoles(member.Roles)
		if member.InviteStatus == "pending" {
			projectDetails.Invited = append(projectDetails.Invited, *member)
		} else if member.InviteStatus == "accepted" {
			projectDetails.Members = append(projectDetails.Members, *member)
		}
	}
	return projectDetails, nil
}

func (s *ProjectService) SearchForUsers(ctx context.Context, userName string) ([]project.ProjectMembership, error) {
	users, err := s.projRepo.SearchForUsers(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("error searching for users: %v", err)
	}
	foundMembers := []project.ProjectMembership{}
	for _, user := range users {
		foundMembers = append(foundMembers, *user)
	}
	return foundMembers, nil
}

func (s *ProjectService) GetProjectUser(ctx context.Context, projectId int, userId int) (*project.ProjectMembership, error) {
	user, err := s.projRepo.GetProjectUser(ctx, projectId, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting user from project: %v", err)
	}
	return user, nil
}

func (s *ProjectService) InviteMember(ctx context.Context, projectId int, userId int) (*project.ProjectMembership, error) {
	// check if member is already invited
	user, err := s.projRepo.GetProjectUser(ctx, projectId, userId)
	if err != nil && err != project.ErrMemberNotFound {
		return nil, fmt.Errorf("error getting user from project: %v", err)
	}
	if err == project.ErrMemberNotFound {
		err = s.projRepo.InviteMember(ctx, projectId, userId)
		if err != nil {
			return nil, fmt.Errorf("error inviting user to project: %v", err)
		}
	}
	if user != nil {
		fmt.Println("user: ", user)
		return user, nil
	}
	invited, err := s.projRepo.GetProjectUser(ctx, projectId, userId)
	fmt.Println("invited: ", invited)
	if err != nil {
		return nil, fmt.Errorf("error getting project members: %v", err)
	}
	return invited, nil
}

// should this be in the user service??
func (s *ProjectService) JoinProject(ctx context.Context, projectId int, userId int, role string) ([]*project.ProjectOverview, error) {
	err := s.projRepo.JoinProject(ctx, projectId, userId)
	if err != nil {
		return nil, fmt.Errorf("error joining project: %v", err)
	}
	projects, err := s.projRepo.GetProjectsForUserSelection(ctx, userId)
	// users, err := s.projRepo.GetProjectUsers(ctx, projectId)
	return projects, nil
}
