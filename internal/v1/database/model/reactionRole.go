package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReactionRole struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	GuildID   string    `gorm:"not null;uniqueIndex:idx_guild_msg_emoji"`
	ChannelID string    `gorm:"not null"`
	MessageID string    `gorm:"not null;uniqueIndex:idx_guild_msg_emoji"`
	EmojiName string    `gorm:"not null;uniqueIndex:idx_guild_msg_emoji"`
	RoleID    string    `gorm:"not null"`
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
