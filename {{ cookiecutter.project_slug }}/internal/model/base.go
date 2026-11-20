package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// BaseModel provides common fields for all models.
// Embed this struct in all domain models.
type BaseModel struct {
	bun.BaseModel `bun:"table:base,alias:b"`

	ID        uuid.UUID  `bun:"type:uuid,pk" json:"id"`
	CreatedAt time.Time  `bun:"type:timestamptz,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt *time.Time `bun:"type:timestamptz,soft_delete" json:"-"`
}

// BeforeInsert sets a UUID if one is not already set.
func (b *BaseModel) BeforeInsert(ctx context.Context) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate updates the timestamp on every update.
func (b *BaseModel) BeforeUpdate(ctx context.Context) error {
	b.UpdatedAt = time.Now()
	return nil
}
