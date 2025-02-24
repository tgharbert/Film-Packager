package projectservice

import (
	"filmPackager/internal/domain/document"
	"filmPackager/internal/domain/membership"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"

	"github.com/google/uuid"
)

func (rv *GetProjectDetailsResponse) sortStagedLockedDocs(documents []*document.Document) {
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

func (rv *GetProjectDetailsResponse) sortMembersByPendingAccepted(members []membership.Membership, users []user.User, userID uuid.UUID) {
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

func (rv *GetUsersProjectsResponse) sortProjectsByPendingAccepted(projects []project.Project, userMemberships []membership.Membership) {
	// colate the projects and memberships on ID
	for _, p := range projects {
		for _, m := range userMemberships {
			if p.ID == m.ProjectID {
				// sort the roles in each project overview here
				m.Roles = membership.SortRoles(m.Roles)
				po := ProjectOverview{
					ID:     p.ID,
					Name:   p.Name,
					Status: m.InviteStatus,
					Roles:  m.Roles,
				}
				// sort them based on invite status
				if m.InviteStatus == "pending" {
					rv.Invited = append(rv.Invited, po)
				} else if m.InviteStatus == "accepted" {
					rv.Accepted = append(rv.Accepted, po)
				}
			}
		}
	}
}
