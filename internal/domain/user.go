package domain

type User struct {
	Id          int
	Name        string
	Email       string
	Password    string
	Memberships []ProjectOverview
	Invited     []ProjectOverview
}
