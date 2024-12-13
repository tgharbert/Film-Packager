package membershipservice

import (
	"context"
	"filmPackager/internal/domain/membership"
	"filmPackager/internal/domain/user"
	"fmt"

	"github.com/google/uuid"
)

type MembershipService struct {
	memberRepo membership.MembershipRepository
	userRepo   user.UserRepository
}

func NewMembershipService(memberRepo membership.MembershipRepository, userRepo user.UserRepository) *MembershipService {
	return &MembershipService{memberRepo: memberRepo, userRepo: userRepo}
}

// search for memberhips based on a term string and a project id
func (s *MembershipService) SearchForNewMembers(ctx context.Context, term string, projectID uuid.UUID) ([]user.User, error) {
	// get all memberships for the project - could write func to get only the user ids
	memberships, err := s.memberRepo.GetProjectMemberships(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting project memberships: %v", err)
	}

	memUserIDs := []uuid.UUID{}

	// create an array of user ids from the memberships
	for _, m := range memberships {
		memUserIDs = append(memUserIDs, m.UserID)
	}

	// get all users with the term in their name and not among the userIDs array
	users, err := s.userRepo.GetAllNewUsersByTerm(ctx, term, memUserIDs)
	if err != nil {
		fmt.Println("error getting users by term: ", err)
		return nil, fmt.Errorf("error getting users by term: %v", err)
	}

	return users, nil
}
