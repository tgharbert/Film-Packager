package project

import (
	"context"

	"github.com/google/uuid"
)

type ProjectRepository interface {
	//	GetProjectsByUserID(ctx context.Context, userId uuid.UUID) ([]Project, error)
	GetProjectsByMembershipIDs(ctx context.Context, memIDs []uuid.UUID) ([]Project, error)
	CreateNewProject(ctx context.Context, project *Project, userId uuid.UUID) error
	DeleteProject(ctx context.Context, projectId uuid.UUID) error
	GetProjectDetails(ctx context.Context, projectId uuid.UUID) (*Project, error)
	//	GetProjectUsers(ctx context.Context, projectId uuid.UUID) ([]ProjectMembership, error)
	//	SearchForUsers(ctx context.Context, userName string) ([]ProjectMembership, error)
	InviteMember(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error
	JoinProject(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error
	//	GetProjectUser(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) (*ProjectMembership, error)
	UpdateMemberRoles(ctx context.Context, projectId uuid.UUID, userId uuid.UUID, role string) error
}
