package permission

import (
	"slices"

	"github.com/SkinonikS/discord-bot-go/internal/v1/util"
)

type Collection struct {
	list []int64
}

func New(perms ...int64) *Collection {
	return &Collection{list: perms}
}

func (s *Collection) Add(perms ...int64) {
	s.list = append(s.list, perms...)
}

func (s *Collection) Has(perm int64) bool {
	return slices.Contains(s.list, perm)
}

func (s *Collection) List() []int64 {
	return s.list
}

func (s *Collection) Remove(perm int64) {
	filtered := make([]int64, 0, len(s.list))
	for _, p := range s.list {
		if p != perm {
			filtered = append(filtered, p)
		}
	}
	s.list = filtered
}

func (s *Collection) AsBit() int64 {
	var bits int64
	for _, p := range s.list {
		bits |= p
	}
	return bits
}

func (s *Collection) AsBitPtr() *int64 {
	return util.ToPtr(s.AsBit())
}
