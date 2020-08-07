package main

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Config provides config options
type Config struct {
	AuthToken     string     `toml:"auth_token"`
	Discriminator string     `toml:"discriminator"`
	AllowedRoles  []RoleDesc `toml:"allowed_roles"`
	RoleChannel   string     `toml:"role_channel"`
	Color         string     `toml:"color"`
}

// RoleDesc is a role tag and description
type RoleDesc struct {
	Role string `toml:"role" json:"role"`
	Desc string `toml:"desc" json:"desc"`
}

// ReadConfig parses a configuration from the given file.
func ReadConfig(filename string) (*Config, error) {
	var c Config

	if _, err := toml.DecodeFile(filename, &c); err != nil {
		return nil, errors.Wrap(err, "error decoding TOML")
	}

	return &c, nil
}
