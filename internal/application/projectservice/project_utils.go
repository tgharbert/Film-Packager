package projectservice

import (
	"filmPackager/internal/domain/document"
	"filmPackager/internal/domain/membership"
	"filmPackager/internal/domain/user"

	"github.com/google/uuid"
)

func (rv *GetProjectDetailsResponse) sortStaged(documents []*document.Document) {
	// make the maps for staged and locked documents
	stagedMap := make(map[string]DocOverview)
	lockedMap := make(map[string]DocOverview)

	// sort the projects by staged or not
	for _, d := range documents {
		// format the document date
		dOverview := &DocOverview{
			ID:   d.ID,
			Date: d.Date.Format("01-02-2006"),
		}
		if d.Status == "staged" {
			// set the bool for staged if there is one
			rv.HasStaged = true
			// assign the document to the map based on the fileType
			stagedMap[d.FileType] = *dOverview
		} else {
			// set the bool for locked if there is one
			rv.HasLocked = true
			// assign the document to the map based on the fileType
			lockedMap[d.FileType] = *dOverview
		}
	}

	// assign the maps to the response
	rv.Staged = &stagedMap
	rv.Locked = &lockedMap
}

func (rv *GetProjectDetailsResponse) sortMembers(members []membership.Membership, users []user.User, userID uuid.UUID) {
	// make a map of userIDs to user data for quicker access
	uMap := make(map[uuid.UUID]user.User)

	// loop through the users and add them to the map
	for _, u := range users {
		uMap[u.Id] = u
	}

	// loop through the members and add the user data
	for _, m := range members {
		m.UserName = uMap[m.UserID].Name
		m.UserEmail = uMap[m.UserID].Email

		// sort the roles
		m.Roles = membership.SortRoles(m.Roles)

		// if the user has the correct status allow them to lock the docs
		if m.HasLockingStatus() && m.UserID == userID {
			rv.LockStatus = true
		}

		if m.Roles[0] != "reader" && m.UserID == userID {
			rv.UploadStatus = true
		}

		// sort the members based on invite status
		if m.InviteStatus == "pending" {
			rv.Invited = append(rv.Invited, m)
		} else if m.InviteStatus == "accepted" {
			rv.Members = append(rv.Members, m)
		}
	}
}
