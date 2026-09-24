package foundation

import "slices"

type RunMode string

const (
	RunModeApplication RunMode = "application"
	RunModeCLI         RunMode = "cli"
)

func (r RunMode) Is(other string) bool {
	return string(r) == other
}

func (r RunMode) IsNot(other string) bool {
	return string(r) != other
}

func (r RunMode) IsOneOf(others ...string) bool {
	return slices.ContainsFunc(others, func(other string) bool {
		return r.Is(other)
	})
}

func (r RunMode) String() string {
	return string(r)
}
