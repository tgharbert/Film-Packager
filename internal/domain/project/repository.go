package project

import "context"

type ProjectRepository interface {
	GetProjectsForUserSelection(ctx context.Context, userId int) ([]*ProjectOverview, error)
	// GetProjectById(ctx, projectId int) (*project.Project, error)
	CreateNewProject(ctx context.Context, projectName string, userId int) (*ProjectOverview, error)
	DeleteProject(ctx context.Context, projectId int) error
	GetProjectDetails(ctx context.Context, projectId int) (*Project, error)
	GetProjectUsers(ctx context.Context, projectId int) ([]*ProjectMembership, error)
	SearchForUsers(ctx context.Context, userName string) ([]*ProjectMembership, error)
	// should be on the user Repository??
	InviteUserToProject(ctx context.Context, projectId int, userId int, role string) error
}
