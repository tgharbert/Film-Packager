// change to projectservice
package projectservice

// package application

import (
	"context"
	"filmPackager/internal/domain/document"
	"filmPackager/internal/domain/membership"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"reflect"
	"time"

	"fmt"

	"github.com/google/uuid"
)

type ProjectService struct {
	projRepo   project.ProjectRepository
	docRepo    document.DocumentRepository
	userRepo   user.UserRepository
	memberRepo membership.MembershipRepository
}

func NewProjectService(projRepo project.ProjectRepository, docRepo document.DocumentRepository, userRepo user.UserRepository, memberRepo membership.MembershipRepository) *ProjectService {
	return &ProjectService{
		projRepo:   projRepo,
		docRepo:    docRepo,
		userRepo:   userRepo,
		memberRepo: memberRepo,
	}
}

type GetProjectDetailsResponse struct {
	Project *project.Project
	Staged  map[string]document.Document
	Locked  map[string]document.Document
	Members []membership.Membership
	Invited []membership.Membership
}

type ProjectOverview struct {
	ID     uuid.UUID
	Name   string
	Status string
	Roles  []string
}

type GetUsersProjectsResponse struct {
	Invited  []ProjectOverview
	Accepted []ProjectOverview
	User     user.User
}

func (s *ProjectService) GetUsersProjects(ctx context.Context, user *user.User) (*GetUsersProjectsResponse, error) {
	rv := &GetUsersProjectsResponse{}
	// do auth work here?
	// get project IDs - new function in repo
	userMemberships, err := s.memberRepo.GetAllUserMemberships(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting user memberships: %v", err)
	}
	projIDs := []uuid.UUID{}
	for _, membership := range userMemberships {
		projIDs = append(projIDs, membership.ProjectID)
	}
	// get the projects for the user
	projects, err := s.projRepo.GetProjectsByMembershipIDs(ctx, projIDs)
	if err != nil {
		return nil, fmt.Errorf("error getting projects from db: %v", err)
	}
	// colate the projects and memberships on ID
	for _, p := range projects {
		for _, m := range userMemberships {
			if p.ID == m.ProjectID {
				// sort the roles in each project overview here
				m.Roles = membership.SortRoles(m.Roles)
				po := ProjectOverview{
					ID:     p.ID,
					Name:   p.Name,
					Status: m.InviteStatus,
					Roles:  m.Roles,
				}
				// sort them based on invite status
				if m.InviteStatus == "pending" {
					rv.Invited = append(rv.Invited, po)
				} else if m.InviteStatus == "accepted" {
					rv.Accepted = append(rv.Accepted, po)
				}
			}
		}
	}
	// get the user info
	user, err = s.userRepo.GetUserById(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting user from db: %v", err)
	}
	rv.User = *user
	return rv, nil
}

func (s *ProjectService) CreateNewProject(ctx context.Context, projectName string, userId uuid.UUID) (*project.ProjectOverview, error) {
	rv := &project.ProjectOverview{}

	u, err := s.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting user from db: %v", err)
	}

	// create the new project with the given info of projectName and userId
	createdProject := &project.Project{
		ID:           uuid.New(),
		Name:         projectName,
		CreatedAt:    time.Now(),
		OwnerID:      userId,
		LastUpdateAt: time.Now(),
	}
	err = s.projRepo.CreateNewProject(ctx, createdProject, userId)

	newMember := &membership.Membership{
		ID:           uuid.New(),
		ProjectID:    createdProject.ID,
		UserID:       userId,
		UserName:     u.Name,
		UserEmail:    u.Email,
		InviteStatus: "accepted",
		Roles:        []string{"owner"},
	}
	if err != nil {
		return nil, fmt.Errorf("error with project creation: %v", err)
	}

	// create a new membership for the owner
	err = s.memberRepo.CreateMembership(ctx, newMember)
	if err != nil {
		return nil, fmt.Errorf("error creating membership: %v", err)
	}
	rv.ID = createdProject.ID
	rv.Name = createdProject.Name
	rv.Status = "invited"
	rv.Roles = newMember.Roles
	return rv, nil
}

// should this be in the user service??
func (s *ProjectService) DeleteProject(ctx context.Context, projectId uuid.UUID, user *user.User) (*GetUsersProjectsResponse, error) {
	rv := &GetUsersProjectsResponse{}
	err := s.projRepo.DeleteProject(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error deleting projects from db: %v", err)
	}
	userMemberships, err := s.memberRepo.GetProjectMemberships(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting user memberships: %v", err)
	}
	projIDs := []uuid.UUID{}
	for _, membership := range userMemberships {
		projIDs = append(projIDs, membership.ProjectID)
	}

	// get the projects for the user
	projects, err := s.projRepo.GetProjectsByMembershipIDs(ctx, projIDs)
	if err != nil {
		return nil, fmt.Errorf("error getting projects from db: %v", err)
	}
	// colate the projects and memberships on ID
	for _, p := range projects {
		for _, m := range userMemberships {
			if p.ID == m.ProjectID {
				// sort the roles in each project overview here
				m.Roles = membership.SortRoles(m.Roles)
				po := ProjectOverview{
					ID:     p.ID,
					Name:   p.Name,
					Status: m.InviteStatus,
					Roles:  m.Roles,
				}
				// sort them based on invite status
				if m.InviteStatus == "pending" {
					rv.Invited = append(rv.Invited, po)
				} else if m.InviteStatus == "accepted" {
					rv.Accepted = append(rv.Accepted, po)
				}
			}
		}
	}
	// get the user info
	user, err = s.userRepo.GetUserById(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting user from db: %v", err)
	}
	rv.User = *user
	return rv, nil
}

func (s *ProjectService) GetProjectDetails(ctx context.Context, projectId uuid.UUID) (*GetProjectDetailsResponse, error) {
	// get the project from the db
	rv := &GetProjectDetailsResponse{}
	p, err := s.projRepo.GetProjectDetails(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project from db: %v", err)
	}
	rv.Project = p
	// get the project documents from the db
	documents, err := s.docRepo.GetAllByOrgId(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project documents from db: %v", err)
	}
	// sort the projects by staged or not
	for _, doc := range documents {
		if doc.IsStaged() {
			setField(&rv.Staged, doc.FileType, doc)
		} else {
			setField(&rv.Locked, doc.FileType, doc)
		}
	}

	// get project members
	members, err := s.memberRepo.GetProjectMemberships(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project users from db: %v", err)
	}
	for _, member := range members {
		member.Roles = membership.SortRoles(member.Roles)
		if member.InviteStatus == "pending" {
			rv.Invited = append(rv.Invited, member)
		} else if member.InviteStatus == "accepted" {
			rv.Members = append(rv.Members, member)
		}
	}
	return rv, nil
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

// MOVE TO MEMBERSHIP SERVICE??
func (s *ProjectService) SearchForUsers(ctx context.Context, userName string) ([]user.User, error) {
	users, err := s.userRepo.GetAllUsersByName(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("error searching for users: %v", err)
	}
	foundMembers := []user.User{}
	for _, user := range users {
		foundMembers = append(foundMembers, user)
	}
	return foundMembers, nil
}

// MOVE TO MEMBERSHIP SERVICE func (s *ProjectService) GetProjectUser(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) (*project.ProjectMembership, error) { user, err := s.projRepo.GetProjectUser(ctx, projectId, userId) sort the roles here user.Roles = project.SortRoles(user.Roles) if err != nil { return nil, fmt.Errorf("error getting user from project: %v", err)}return user, nil}

func (s *ProjectService) InviteMember(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) ([]membership.Membership, error) {
	// check if member is already invited
	user, err := s.memberRepo.GetMembership(ctx, projectId, userId)
	if err != nil && err != project.ErrMemberNotFound {
		return nil, fmt.Errorf("error getting user from project: %v", err)
	}
	if err == project.ErrMemberNotFound {
		err = s.projRepo.InviteMember(ctx, projectId, userId)
		if err != nil {
			return nil, fmt.Errorf("error inviting user to project: %v", err)
		}
	}
	if user.UserName != "" && user.InviteStatus == "pending" {
		return nil, project.ErrMemberAlreadyInvited
	}
	members, err := s.memberRepo.GetProjectMemberships(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project members: %v", err)
	}
	membersInfo := []membership.Membership{}
	for _, member := range members {
		if member.InviteStatus == "pending" {
			membersInfo = append(membersInfo, member)
		}
	}
	return membersInfo, nil
}

func (s *ProjectService) JoinProject(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) ([]project.Project, error) {
	err := s.projRepo.JoinProject(ctx, projectId, userId)
	if err != nil {
		return nil, fmt.Errorf("error joining project: %v", err)
	}
	projects := []project.Project{}
	//	projects, err := s.projRepo.GetProjectsByUserID(ctx, userId)
	return projects, nil
}

func (s *ProjectService) UpdateMemberRoles(ctx context.Context, projectId uuid.UUID, memberId uuid.UUID, userId uuid.UUID, role string) (*membership.Membership, error) {
	// check user permissions...
	user, err := s.memberRepo.GetMembership(ctx, projectId, userId)
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
	member, err := s.memberRepo.GetMembership(ctx, projectId, memberId)
	return member, nil
}
