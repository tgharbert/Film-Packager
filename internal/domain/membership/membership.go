package membership

import (
	"slices"

	"github.com/google/uuid"
)

// this should be defined in the application layer
type Membership struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ProjectID uuid.UUID
	UserName  string
	UserEmail string
	// should be enum for roles?
	Roles []string
	// should be enum for invite status?
	InviteStatus string
}

func SortRoles(rolesSlc []string) []string {
	var orderedRoles []string
	if slices.Contains(rolesSlc, "owner") {
		orderedRoles = append(orderedRoles, "owner")
	}
	if slices.Contains(rolesSlc, "director") {
		orderedRoles = append(orderedRoles, "director")
	}
	if slices.Contains(rolesSlc, "producer") {
		orderedRoles = append(orderedRoles, "producer")
	}
	if slices.Contains(rolesSlc, "writer") {
		orderedRoles = append(orderedRoles, "writer")
	}
	if slices.Contains(rolesSlc, "cinematographer") {
		orderedRoles = append(orderedRoles, "cinematographer")
	}
	if slices.Contains(rolesSlc, "production designer") {
		orderedRoles = append(orderedRoles, "production designer")
	}
	if slices.Contains(rolesSlc, "reader") {
		orderedRoles = append(orderedRoles, "reader")
	}
	return orderedRoles
}

func (m *Membership) HasLockingStatus() bool {
	return slices.Contains(m.Roles, "owner") || slices.Contains(m.Roles, "director") || slices.Contains(m.Roles, "producer")
}
