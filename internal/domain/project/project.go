package project

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID           uuid.UUID
	Name         string
	CreatedAt    time.Time
	OwnerID      uuid.UUID
	LastUpdateAt time.Time
}

type ProjectOverview struct {
	ID     uuid.UUID
	Name   string
	Status string
	Roles  []string
}
