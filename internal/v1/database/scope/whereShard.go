package scope

import (
	"fmt"

	"gorm.io/gorm"
)

func ByShard(shardID, shardCount uint32) func(*gorm.Statement) {
	if shardID > shardCount {
		return func(tx *gorm.Statement) {
			tx.Error = fmt.Errorf("shard id %d is out of range", shardID)
		}
	}

	return func(tx *gorm.Statement) {
		tx.Where("(guild_id >> 22) % ? = ?", shardCount, shardID)
	}
}
