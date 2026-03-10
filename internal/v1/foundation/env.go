package foundation

import "slices"

type Env string

func (e Env) Is(other string) bool {
	return string(e) == other
}

func (e Env) IsNot(other string) bool {
	return string(e) != other
}

func (e Env) IsOneOf(others ...string) bool {
	return slices.ContainsFunc(others, func(other string) bool {
		return e.Is(other)
	})
}

func (e Env) IsProduction() bool {
	return e.IsOneOf("production", "prod")
}

func (e Env) IsDevelopment() bool {
	return e.IsOneOf("development", "dev")
}

func (e Env) IsTest() bool {
	return e.IsOneOf("testing", "test")
}

func (e Env) String() string {
	return string(e)
}
