package main

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"github.com/caromo/rinako/collections"
)

// Config provides config options
type Config struct {
	AuthToken     string                 `toml:"auth_token"`
	Discriminator string                 `toml:"discriminator"`
	AllowedRoles  []collections.RoleDesc `toml:"allowed_roles"`
	RoleChannel   string                 `toml:"role_channel"`
	Color         string                 `toml:"color"`
	DBPath        string                 `toml:"db_path"`
}

// ReadConfig parses a configuration from the given file.
func ReadConfig(filename string) (*Config, error) {
	var c Config

	if _, err := toml.DecodeFile(filename, &c); err != nil {
		return nil, errors.Wrap(err, "error decoding TOML")
	}

	return &c, nil
}
