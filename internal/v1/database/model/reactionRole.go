package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReactionRole struct {
	ID        uuid.UUID
	ChannelID string
	MessageID string
	EmojiName string
	RoleID    string
}

func (u *ReactionRole) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (*ReactionRole) TableName() string {
	return "reaction_roles"
}
