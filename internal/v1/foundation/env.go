package foundation

type Env string

func (e Env) Is(other string) bool {
	return string(e) == other
}

func (e Env) IsNot(other string) bool {
	return string(e) != other
}

func (e Env) IsOneOf(others ...string) bool {
	for _, other := range others {
		if e.Is(other) {
			return true
		}
	}

	return false
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
