package foundation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type BuildInfo struct {
	tag       string
	buildTime string
	commit    string
}

func NewBuildInfo(tag, buildTime, commit string) *BuildInfo {
	return &BuildInfo{tag, buildTime, commit}
}

func (b *BuildInfo) Tag() string {
	return b.tag
}

func (b *BuildInfo) BuildTime() string {
	return b.buildTime
}

func (b *BuildInfo) Commit() string {
	return b.commit
}

func (b *BuildInfo) Hash() string {
	data := fmt.Sprintf("%s%s%s", b.tag, b.buildTime, b.commit)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
