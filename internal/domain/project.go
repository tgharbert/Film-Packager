package domain

type Project struct {
	Id      int
	Name    string
	Locked  []Document
	Staged  []Document
	Members []User
	Invited []User
}

type ProjectOverview struct {
	ID     int
	Name   string
	Status string
	Roles  []string
}
