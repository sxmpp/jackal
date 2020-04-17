/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package app

import (
	"bytes"
	"io/ioutil"

	"github.com/sxmpp/jackal/c2s"
	"github.com/sxmpp/jackal/component"
	"github.com/sxmpp/jackal/module"
	"github.com/sxmpp/jackal/router/host"
	"github.com/sxmpp/jackal/s2s"
	"github.com/sxmpp/jackal/storage"
	"gopkg.in/yaml.v2"
)

// debugConfig represents debug server configuration.
type debugConfig struct {
	Port int `yaml:"port"`
}

type loggerConfig struct {
	Level   string `yaml:"level"`
	LogPath string `yaml:"log_path"`
}

// Config represents a global configuration.
type Config struct {
	PIDFile    string           `yaml:"pid_path"`
	Debug      debugConfig      `yaml:"debug"`
	Logger     loggerConfig     `yaml:"logger"`
	Storage    storage.Config   `yaml:"storage"`
	Hosts      []host.Config    `yaml:"hosts"`
	Modules    module.Config    `yaml:"modules"`
	Components component.Config `yaml:"components"`
	C2S        []c2s.Config     `yaml:"c2s"`
	S2S        *s2s.Config      `yaml:"s2s"`
}

// FromFile loads default global configuration from a specified file.
func (cfg *Config) FromFile(configFile string) error {
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, cfg)
}

// FromBuffer loads default global configuration from a specified byte buffer.
func (cfg *Config) FromBuffer(buf *bytes.Buffer) error {
	return yaml.Unmarshal(buf.Bytes(), cfg)
}
