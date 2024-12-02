package project

import (
	"filmPackager/internal/domain/document"
	"slices"
)

type Project struct {
	Id              int
	Name            string
	Locked          ProjectDocs
	Staged          ProjectDocs
	Members         []ProjectMembership
	Invited         []ProjectMembership
	SearchedMembers []ProjectMembership
}

type ProjectMembership struct {
	UserId       int
	UserName     string
	UserEmail    string
	Roles        []string
	InviteStatus string
}

type ProjectOverview struct {
	Id     int
	Name   string
	Status string
	Roles  []string
}

type ProjectDocs struct {
	Script            document.Document
	Logline           document.Document
	Synopsis          document.Document
	PitchDeck         document.Document
	Schedule          document.Document
	Budget            document.Document
	DirectorStatement document.Document
	Shotlist          document.Document
	Lookbook          document.Document
	Bios              document.Document
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
