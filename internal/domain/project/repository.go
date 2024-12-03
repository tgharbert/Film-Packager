package project

import "context"

type ProjectRepository interface {
	GetProjectsForUserSelection(ctx context.Context, userId int) ([]*ProjectOverview, error)
	CreateNewProject(ctx context.Context, projectName string, userId int) (*ProjectOverview, error)
	DeleteProject(ctx context.Context, projectId int) error
	GetProjectDetails(ctx context.Context, projectId int) (*Project, error)
	GetProjectUsers(ctx context.Context, projectId int) ([]*ProjectMembership, error)
	SearchForUsers(ctx context.Context, userName string) ([]*ProjectMembership, error)
	InviteMember(ctx context.Context, projectId int, userId int) error
	JoinProject(ctx context.Context, projectId int, userId int) error
	GetProjectUser(ctx context.Context, projectId int, userId int) (*ProjectMembership, error)
	UpdateMemberRoles(ctx context.Context, projectId int, userId int, role string) error
}
