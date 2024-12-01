package project

import (
	"filmPackager/internal/domain/document"
	"strings"
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
	UserId    int
	UserName  string
	UserEmail string
	Roles     []string
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

func SortRoles(rolesStr string) []string {
	var orderedRoles []string
	if strings.Contains(rolesStr, "owner") {
		orderedRoles = append(orderedRoles, "owner")
	}
	if strings.Contains(rolesStr, "director") {
		orderedRoles = append(orderedRoles, "director")
	}
	if strings.Contains(rolesStr, "producer") {
		orderedRoles = append(orderedRoles, "producer")
	}
	if strings.Contains(rolesStr, "writer") {
		orderedRoles = append(orderedRoles, "writer")
	}
	if strings.Contains(rolesStr, "cinematographer") {
		orderedRoles = append(orderedRoles, "cinematographer")
	}
	if strings.Contains(rolesStr, "production designer") {
		orderedRoles = append(orderedRoles, "production designer")
	}
	if strings.Contains(rolesStr, "reader") {
		orderedRoles = append(orderedRoles, "reader")
	}
	return orderedRoles
}
