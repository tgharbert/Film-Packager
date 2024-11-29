package domain

// do I need to rethink this entity - structs for docs to allow easy access??
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
	Script            Document
	Logline           Document
	Synopsis          Document
	PitchDeck         Document
	Schedule          Document
	Budget            Document
	DirectorStatement Document
	Shotlist          Document
	Lookbook          Document
	Bios              Document
}
