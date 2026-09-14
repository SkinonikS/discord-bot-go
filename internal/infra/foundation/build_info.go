package foundation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type BuildInfo struct {
	Tag       string
	BuildTime string
	Commit    string
}

func (b *BuildInfo) Hash() string {
	data := fmt.Sprintf("%s%s%s", b.Tag, b.BuildTime, b.Commit)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
