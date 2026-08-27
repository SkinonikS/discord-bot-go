package cfgutil

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func CastConfig[T any](cfg any) (T, error) {
	var result T

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return result, fmt.Errorf("failed to marshal: %w", err)
	}

	if err := yaml.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("failed to parse: %w", err)
	}

	return result, nil
}
