package membershipservice

import (
	"context"
	"filmPackager/internal/domain/membership"
	"filmPackager/internal/domain/user"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

type MembershipService struct {
	memberRepo membership.MembershipRepository
	userRepo   user.UserRepository
}

type GetMembershipResponse struct {
	Membership     *membership.Membership
	AvailableRoles []string
}

func NewMembershipService(memberRepo membership.MembershipRepository, userRepo user.UserRepository) *MembershipService {
	return &MembershipService{memberRepo: memberRepo, userRepo: userRepo}
}

type GetProjectMembershipsResponse struct {
	Invited []membership.Membership
	Members []membership.Membership
}

type UpdateMemberRolesResponse struct {
	UserName   string
	Membership *membership.Membership
	Roles      []string
}

func (s *MembershipService) SearchForNewMembersByName(ctx context.Context, name string, projectID uuid.UUID) ([]user.User, error) {
	// remove the whitespace in the name first
	name = strings.Join(strings.Fields(name), "")

	// return an error if the name is below the search threshold
	if len(name) < 3 {
		return nil, membership.ErrSearchTermTooShort
	}

	// get all memberships for the project - could write func to get only the user ids
	memberships, err := s.memberRepo.GetProjectMemberships(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting project memberships: %v", err)
	}

	// create an array of user ids from the memberships
	memUserIDs := []uuid.UUID{}
	for _, m := range memberships {
		memUserIDs = append(memUserIDs, m.UserID)
	}

	// get all users with the term in their name and not among the userIDs array
	users, err := s.userRepo.GetAllNewUsersByName(ctx, name, memUserIDs)
	if err != nil {
		return nil, fmt.Errorf("error getting users by term: %v", err)
	}

	return users, nil
}

// invite a user to a project
func (s *MembershipService) InviteUserToProject(ctx context.Context, userID, projectID uuid.UUID) ([]membership.Membership, error) {
	// get the user by id
	u, err := s.userRepo.GetUserById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user by id: %v", err)
	}

	// check if the user is already a member of the project
	_, err = s.memberRepo.GetMembership(ctx, userID, projectID)
	if err == nil {
		return nil, membership.ErrUserAlreadyMember
	}

	// create the membership
	newMember := &membership.Membership{
		ID:           uuid.New(),
		UserID:       userID,
		UserName:     u.Name,
		UserEmail:    u.Email,
		ProjectID:    projectID,
		InviteStatus: "pending",
		Roles:        []string{"reader"},
	}

	err = s.memberRepo.CreateMembership(ctx, newMember)
	if err != nil {
		return nil, fmt.Errorf("error creating new membership: %v", err)
	}

	// get all memberships for the project
	memberships, err := s.memberRepo.GetProjectMemberships(ctx, projectID)

	// sort them by pending or member
	invited := []membership.Membership{}
	invitedIDs := []uuid.UUID{}

	for _, m := range memberships {
		if m.InviteStatus == "pending" {
			invited = append(invited, m)
			invitedIDs = append(invitedIDs, m.UserID)
		}
	}

	// get the users by the ids
	users, err := s.userRepo.GetUsersByIDs(ctx, invitedIDs)
	if err != nil {
		return nil, fmt.Errorf("error getting users by ids: %v", err)
	}

	// add the names and emails to the memberships
	for _, u := range users {
		for idx, i := range invited {
			if u.Id == i.UserID {
				invited[idx].UserName = u.Name
				invited[idx].UserEmail = u.Email

			}
		}
	}

	return invited, nil
}

// get a user's memberships for a project
func (s *MembershipService) GetMembership(ctx context.Context, projectID, userID uuid.UUID) (*GetMembershipResponse, error) {
	m, err := s.memberRepo.GetMembership(ctx, projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting membership: %v", err)
	}

	u, err := s.userRepo.GetUserById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user by id: %v", err)
	}

	// fill in the user's name and email
	m.UserName = u.Name
	m.UserEmail = u.Email

	// return the roles still available to this member
	var availRoles []string
	allRoles := []string{"director", "producer", "writer", "cinematographer", "production_designer"}
	for _, role := range allRoles {
		if slices.Contains(m.Roles, role) {
			continue
		} else {
			availRoles = append(availRoles, role)
		}
	}

	rv := &GetMembershipResponse{
		Membership:     m,
		AvailableRoles: availRoles,
	}

	return rv, nil
}

func (s *MembershipService) UpdateMemberRoles(ctx context.Context, projectID, userID uuid.UUID, role string) (*membership.Membership, error) {
	// get the membership
	m, err := s.memberRepo.GetMembership(ctx, projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting membership: %v", err)
	}

	u, err := s.userRepo.GetUserById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user by id: %v", err)
	}

	// attach the username to the membership
	m.UserName = u.Name

	// remove the reader role upon addition of further roles
	if m.Roles[0] == "reader" {
		// effectively setting it to an empty slice
		m.Roles = m.Roles[1:]
	}

	// add the role
	m.Roles = append(m.Roles, role)
	m.Roles = membership.SortRoles(m.Roles)

	// update the membership
	err = s.memberRepo.UpdateMembership(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("error updating membership: %v", err)
	}

	return m, nil
}

func (s *MembershipService) GetProjectMemberships(ctx context.Context, projectID uuid.UUID) (*GetProjectMembershipsResponse, error) {
	rv := &GetProjectMembershipsResponse{}

	memberships, err := s.memberRepo.GetProjectMemberships(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting project memberships: %v", err)
	}

	userIDs := []uuid.UUID{}

	// get userIds from memberships
	for _, m := range memberships {
		userIDs = append(userIDs, m.UserID)
	}

	users, err := s.userRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("error getting users by ids: %v", err)
	}

	for _, m := range memberships {
		for _, u := range users {
			// assign the user's name and email to the membership
			if m.UserID == u.Id {
				m.UserName = u.Name
				m.UserEmail = u.Email
			}
		}
		// sort memberships by pending or member
		if m.InviteStatus == "pending" {
			rv.Invited = append(rv.Invited, m)
		} else {
			rv.Members = append(rv.Members, m)
		}
	}

	return rv, nil
}
