package model

import (
	"github.com/google/uuid"
)

// Item represents a domain entity.
type Item struct {
	BaseModel `bun:"table:items,alias:i"`

	Name        string     `bun:"type:varchar(255),notnull" json:"name"`
	Description string     `bun:"type:text" json:"description"`
	CreatedBy   *uuid.UUID `bun:"type:uuid" json:"created_by,omitempty"`
}
