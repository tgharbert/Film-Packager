package application

import (
	"context"
	"filmPackager/internal/domain/document"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"reflect"

	"fmt"
)

type ProjectService struct {
	projRepo project.ProjectRepository
	docRepo  document.DocumentRepository
}

func NewProjectService(projRepo project.ProjectRepository, docRepo document.DocumentRepository) *ProjectService {
	return &ProjectService{
		projRepo: projRepo,
		docRepo:  docRepo,
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

	// get the project documents from the db
	documents, err := s.docRepo.GetAllByOrgId(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project documents from db: %v", err)
	}
	// sort the projects by staged or not
	for _, doc := range documents {
		if doc.IsStaged() {
			setField(&projectDetails.Staged, doc.FileType, doc)
		} else {
			setField(&projectDetails.Locked, doc.FileType, doc)
		}
	}

	// get project members
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

// MOVE TO DOMAIN?
func setField(obj interface{}, fieldName string, value interface{}) {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(fieldName)
	if !structFieldValue.IsValid() {
		return
	}
	if !structFieldValue.CanSet() {
		return
	}
	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType == val.Type() {
		structFieldValue.Set(val)
	}
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
	// sort the roles here
	user.Roles = project.SortRoles(user.Roles)
	if err != nil {
		return nil, fmt.Errorf("error getting user from project: %v", err)
	}
	return user, nil
}

func (s *ProjectService) InviteMember(ctx context.Context, projectId int, userId int) ([]project.ProjectMembership, error) {
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
	if user != nil && user.InviteStatus == "pending" {
		return nil, project.ErrMemberAlreadyInvited
	}
	members, err := s.projRepo.GetProjectUsers(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project members: %v", err)
	}
	membersInfo := []project.ProjectMembership{}
	for _, member := range members {
		if member.InviteStatus == "pending" {
			membersInfo = append(membersInfo, *member)
		}
	}
	return membersInfo, nil
}

func (s *ProjectService) JoinProject(ctx context.Context, projectId int, userId int) ([]*project.ProjectOverview, error) {
	err := s.projRepo.JoinProject(ctx, projectId, userId)
	if err != nil {
		return nil, fmt.Errorf("error joining project: %v", err)
	}
	projects, err := s.projRepo.GetProjectsForUserSelection(ctx, userId)
	return projects, nil
}

func (s *ProjectService) UpdateMemberRoles(ctx context.Context, projectId int, memberId int, userId int, role string) (*project.ProjectMembership, error) {
	// check user permissions...
	user, err := s.projRepo.GetProjectUser(ctx, projectId, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting user from project: %v", err)
	}
	// TO ADD: establish slice of roles that allow for role updates
	if user.Roles[0] != "owner" {
		return nil, fmt.Errorf("user does not have permission to update roles")
	}
	err = s.projRepo.UpdateMemberRoles(ctx, projectId, memberId, role)
	if err != nil {
		return nil, fmt.Errorf("error updating member roles: %v", err)
	}
	member, err := s.projRepo.GetProjectUser(ctx, projectId, memberId)
	return member, nil
}
