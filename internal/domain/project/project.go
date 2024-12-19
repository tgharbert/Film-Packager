package project

import (
	"time"

	"github.com/google/uuid"
)

// TODO: DELETE ALL COMMENTS AND MOVE STRUCTS TO RESPONSES IN APPLICATION LAYER
type Project struct {
	ID           uuid.UUID
	Name         string
	CreatedAt    time.Time
	OwnerID      uuid.UUID
	LastUpdateAt time.Time
	//	Locked ProjectDocs
	//	Staged ProjectDocs
	// Members         []ProjectMembership
	// Invited         []ProjectMembership
	// SearchedMembers []ProjectMembership
}

// this should be defined in the application layer
//type ProjectMembership struct {
//	UserID       int
//	UserName     string
//	UserEmail    string
//	Roles        []string
//	InviteStatus string
//}

// should this take in the User then get the projects and sort them for the user??
type ProjectOverview struct {
	ID     uuid.UUID
	Name   string
	Status string
	Roles  []string
}

type GetUsersProjects struct {
	// should be a separate call in the application layer
	// User     *user.User
	Pending  []ProjectOverview
	Accepted []ProjectOverview
}

//type GetProjectDetails struct {
//	Project *Project
// these should be separate calls in the application layer
//Staged  []document.Document
//Locked  []document.Document
//	Members []ProjectMembership
//	Invited []ProjectMembership
//}

// this should be defined in the application layer

//type ProjectDocs struct {
//	Script            *document.Document
//	Logline           *document.Document
//	Synopsis          *document.Document
//	PitchDeck         *document.Document
//	Schedule          *document.Document
//	Budget            *document.Document
//	DirectorStatement *document.Document
//	Shotlist          *document.Document
//	Lookbook          *document.Document
//	Bios              *document.Document
//}
