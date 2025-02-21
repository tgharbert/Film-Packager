package project

import (
	"context"

	"github.com/google/uuid"
)

type ProjectRepository interface {
	GetProjectsByMembershipIDs(ctx context.Context, memIDs []uuid.UUID) ([]Project, error)
	CreateNewProject(ctx context.Context, project *Project, userId uuid.UUID) error
	DeleteProject(ctx context.Context, projectId uuid.UUID) error
	GetProjectByID(ctx context.Context, projectId uuid.UUID) (*Project, error)
	InviteMember(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error
	JoinProject(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error
	UpdateMemberRoles(ctx context.Context, projectId uuid.UUID, userId uuid.UUID, role string) error
	UpdateProject(ctx context.Context, project *Project) error
}
