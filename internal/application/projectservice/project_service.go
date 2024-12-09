// change to projectservice
package projectservice

// package application

import (
	"context"
	"filmPackager/internal/domain/document"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"reflect"

	"fmt"

	"github.com/google/uuid"
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
type ProjectOverview struct {
	Id     int
	Name   string
	Status string
	Roles  []string
}

type GetUsersProjects struct {
	User     *user.User
	Pending  []project.ProjectOverview
	Accepted []project.ProjectOverview
}

type GetProjectDetails struct {
	Project *project.Project
	Staged  []document.Document
	Locked  []document.Document
	Members []project.ProjectMembership
	Invited []project.ProjectMembership
}

func (s *ProjectService) GetUsersProjects(ctx context.Context, user *user.User) (*GetUsersProjects, error) {
	var rv *GetUsersProjects
	// do auth work here?
	projects, err := s.projRepo.GetProjectsForUserSelection(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting projects from db: %v", err)
	}
	for _, project := range projects {
		// sort the roles in each project here as well
		if project.Status == "pending" {
			rv.Pending = append(rv.Pending, *project)
		}
		if project.Status == "accepted" {
			rv.Accepted = append(rv.Accepted, *project)
		}
	}
	return rv, nil
}

func (s *ProjectService) CreateNewProject(ctx context.Context, projectName string, userId uuid.UUID) (*project.ProjectOverview, error) {
	// create the new project with the given info of projectName and userId
	project, err := s.projRepo.CreateNewProject(ctx, projectName, userId)
	if err != nil {
		return nil, fmt.Errorf("error with project creation: %v", err)
	}
	return project, nil
}

// should this be in the user service??
func (s *ProjectService) DeleteProject(ctx context.Context, projectId uuid.UUID, user *user.User) (*GetUsersProjects, error) {
	var rv *GetUsersProjects
	err := s.projRepo.DeleteProject(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error deleting projects from db: %v", err)
	}
	projects, err := s.projRepo.GetProjectsForUserSelection(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user projects: %v", err)
	}
	for _, project := range projects {
		// sort the roles in each project here as well
		if project.Status == "pending" {
			rv.Pending = append(rv.Pending, *project)
		}
		if project.Status == "accepted" {
			rv.Accepted = append(rv.Accepted, *project)
		}
	}
	return rv, nil
}

func (s *ProjectService) GetProjectDetails(ctx context.Context, projectId uuid.UUID) (*GetProjectDetails, error) {
	var projectDetails *GetProjectDetails
	// get the project from the db
	p, err := s.projRepo.GetProjectDetails(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project from db: %v", err)
	}
	projectDetails.Project = p
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
