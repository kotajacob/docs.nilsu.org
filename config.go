package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Address      string
	ReadTimeout  int64
	WriteTimeout int64
	TemplateDir  string
	StaticDir    string
	DataDir      string
	HashDir      string
}

func (c *Config) Load(p string) error {
	if _, err := toml.DecodeFile(p, c); err != nil {
		return err
	}
	return nil
}
