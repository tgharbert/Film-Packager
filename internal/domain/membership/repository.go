package membership

import (
	"context"

	"github.com/google/uuid"
)

type MembershipRepository interface {
	GetProjectMemberships(ctx context.Context, projectId uuid.UUID) ([]Membership, error)
	CreateMembership(ctx context.Context, membership *Membership) error
	DeleteMembership(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error
	GetUserMembershipsForProject(ctx context.Context, userId uuid.UUID, projectId uuid.UUID) (*Membership, error)
	GetMembership(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) (*Membership, error)
	GetProjectIDsForUser(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error)
	GetAllUserMemberships(ctx context.Context, userId uuid.UUID) ([]Membership, error)
}
