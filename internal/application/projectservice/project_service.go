package projectservice

import (
	"context"
	"filmPackager/internal/domain/document"
	"filmPackager/internal/domain/membership"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"slices"
	"time"

	"fmt"

	"github.com/google/uuid"
)

type ProjectService struct {
	projRepo   project.ProjectRepository
	docRepo    document.DocumentRepository
	s3Repo     document.S3Repository
	userRepo   user.UserRepository
	memberRepo membership.MembershipRepository
}

func NewProjectService(projRepo project.ProjectRepository, docRepo document.DocumentRepository, s3Repo document.S3Repository, userRepo user.UserRepository, memberRepo membership.MembershipRepository) *ProjectService {
	return &ProjectService{
		projRepo:   projRepo,
		docRepo:    docRepo,
		s3Repo:     s3Repo,
		userRepo:   userRepo,
		memberRepo: memberRepo,
	}
}

type GetProjectDetailsResponse struct {
	Project      *project.Project
	Staged       *map[string]DocOverview
	Locked       *map[string]DocOverview
	Members      []membership.Membership
	Invited      []membership.Membership
	LockStatus   bool
	UploadStatus bool
	HasLocked    bool
	HasStaged    bool
}

type DocOverview struct {
	ID   uuid.UUID
	Date string
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

func (s *ProjectService) GetProject(ctx context.Context, pID uuid.UUID) (*project.Project, error) {
	p, err := s.projRepo.GetProjectByID(ctx, pID)
	if err != nil {
		return nil, fmt.Errorf("error getting project from db: %v", err)
	}

	return p, nil
}

func (s *ProjectService) GetProjectOverview(ctx context.Context, pID uuid.UUID, uID uuid.UUID) (*ProjectOverview, error) {
	rv := &ProjectOverview{}
	// get the project
	p, err := s.projRepo.GetProjectByID(ctx, pID)
	if err != nil {
		return nil, fmt.Errorf("error getting project from db: %v", err)
	}

	// get the user membership
	m, err := s.memberRepo.GetMembership(ctx, pID, uID)
	if err != nil {
		return nil, fmt.Errorf("error getting user membership: %v", err)
	}

	// sort the roles
	m.Roles = membership.SortRoles(m.Roles)

	// assign the project overview
	rv.ID = p.ID
	rv.Name = p.Name
	rv.Status = m.InviteStatus
	rv.Roles = m.Roles

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

func (s *ProjectService) DeleteProject(ctx context.Context, projectId uuid.UUID, user *user.User) (*GetUsersProjectsResponse, error) {
	rv := &GetUsersProjectsResponse{}
	// function to get all of the documents for a project
	docs, err := s.docRepo.GetAllByOrgId(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project documents from db: %v", err)
	}

	// put all project IDs in a slice
	keys := []string{}
	for _, d := range docs {
		k := fmt.Sprintf("%s=%s", d.FileName, d.ID)
		keys = append(keys, k)
	}

	// delete all the project files from s3
	err = s.s3Repo.DeleteAllOrgFiles(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("error deleting project files from s3: %v", err)
	}

	// delete the project from the db
	err = s.projRepo.DeleteProject(ctx, projectId)
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

func (s *ProjectService) GetProjectDetails(ctx context.Context, projectId uuid.UUID, userID uuid.UUID) (*GetProjectDetailsResponse, error) {
	// get the project from the db
	rv := &GetProjectDetailsResponse{}

	// get project details
	p, err := s.projRepo.GetProjectByID(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project from db: %v", err)
	}

	// assign the project to the response
	rv.Project = p

	// get the project documents from the db
	documents, err := s.docRepo.GetAllByOrgId(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project documents from db: %v", err)
	}

	// set bools for staged and locked documents
	rv.HasLocked = false
	rv.HasStaged = false

	// make the maps for staged and locked documents
	stagedMap := make(map[string]DocOverview)
	lockedMap := make(map[string]DocOverview)

	// sort the projects by staged or not
	for _, d := range documents {
		// format the document date
		dOverview := &DocOverview{
			ID:   d.ID,
			Date: d.Date.Format("01-02-2006"),
		}
		if d.Status == "staged" {
			// set the bool for staged if there is one
			rv.HasStaged = true
			// assign the document to the map based on the fileType
			stagedMap[d.FileType] = *dOverview
		} else {
			// set the bool for locked if there is one
			rv.HasLocked = true
			// assign the document to the map based on the fileType
			lockedMap[d.FileType] = *dOverview
		}
	}

	// assign the maps to the response
	rv.Staged = &stagedMap
	rv.Locked = &lockedMap

	// get project members
	members, err := s.memberRepo.GetProjectMemberships(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project users from db: %v", err)
	}

	// build an array of member userIDs
	mIDs := []uuid.UUID{}
	for _, m := range members {
		mIDs = append(mIDs, m.UserID)
	}

	// make a map of userIDs to user data for quicker access
	uMap := make(map[uuid.UUID]user.User)

	// get the user data
	users, err := s.userRepo.GetUsersByIDs(ctx, mIDs)
	for _, u := range users {
		uMap[u.Id] = u
	}

	rv.LockStatus = false
	rv.UploadStatus = false

	// loop through the members and add the user data
	for _, m := range members {
		m.UserName = uMap[m.UserID].Name
		m.UserEmail = uMap[m.UserID].Email

		// sort the roles
		m.Roles = membership.SortRoles(m.Roles)

		// if the user has the correct status allow them to lock the docs
		if memberHasLockingStatus(m.Roles) && m.UserID == userID {
			rv.LockStatus = true
		}
		if m.Roles[0] != "reader" && m.UserID == userID {
			rv.UploadStatus = true
		}

		// sort the members based on invite status
		if m.InviteStatus == "pending" {
			rv.Invited = append(rv.Invited, m)
		} else if m.InviteStatus == "accepted" {
			rv.Members = append(rv.Members, m)
		}
	}

	return rv, nil
}

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

func (s *ProjectService) JoinProject(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error {
	err := s.projRepo.JoinProject(ctx, projectId, userId)
	if err != nil {
		return fmt.Errorf("error joining project: %v", err)
	}

	return nil
}

func (s *ProjectService) UpdateProjectName(ctx context.Context, projectId uuid.UUID, newName string) (*project.Project, error) {
	p, err := s.projRepo.GetProjectByID(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project: %v", err)
	}

	p.Name = newName
	p.LastUpdateAt = time.Now()

	err = s.projRepo.UpdateProject(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("error editing project name: %v", err)
	}

	updatedP, err := s.projRepo.GetProjectByID(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting project: %v", err)
	}

	return updatedP, nil
}

// utility functions
func memberHasLockingStatus(roles []string) bool {
	if slices.Contains(roles, "owner") || slices.Contains(roles, "director") || slices.Contains(roles, "producer") {
		return true
	}
	return false
}
