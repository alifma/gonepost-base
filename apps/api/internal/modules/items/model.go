// Package items is the reference CRUD module. Copy it to start a new
// feature — docs/guides/adding-a-feature.md walks through every step.
package items

import (
	"time"

	"basecode/api/internal/platform/openapigen"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

// Item is the internal representation. Every row belongs to one owner.
type Item struct {
	ID          string
	OwnerID     string
	Name        string
	Description *string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Input is the writable part of an Item (create and update).
type Input struct {
	Name        string
	Description *string
	Status      Status
}

// Public converts to the OpenAPI-generated Item type, so a contract change
// fails to compile here until it is matched.
func (i Item) Public() openapigen.Item {
	return openapigen.Item{
		Id:          i.ID,
		Name:        i.Name,
		Description: i.Description,
		Status:      openapigen.ItemStatus(i.Status),
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}
