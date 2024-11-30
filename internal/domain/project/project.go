package project

import (
	"filmPackager/internal/domain/document"
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
