package config

import (
	"github.com/kelseyhightower/envconfig"
)

// NewConfig generates new app configuration for type, given by generics.
func New(prefix ...string) (*Config, error) {
	var v Config
	p := ""
	if len(prefix) > 0 {
		p = prefix[0]
	}

	if err := envconfig.Process(p, &v); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &v, nil
}
