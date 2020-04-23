/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package module

import (
	"fmt"

	"github.com/sxmpp/jackal/module/offline"
	"github.com/sxmpp/jackal/module/roster"
	"github.com/sxmpp/jackal/module/xep0092"
	"github.com/sxmpp/jackal/module/xep0199"
)

// Config represents C2S modules configuration.
type Config struct {
	Enabled map[string]struct{}
	Roster  roster.Config
	Offline offline.Config
	Version xep0092.Config
	Ping    xep0199.Config
}

type configProxy struct {
	Enabled []string       `yaml:"enabled"`
	Roster  roster.Config  `yaml:"mod_roster"`
	Offline offline.Config `yaml:"mod_offline"`
	Version xep0092.Config `yaml:"mod_version"`
	Ping    xep0199.Config `yaml:"mod_ping"`
}

// UnmarshalYAML satisfies Unmarshaler interface.
func (cfg *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	p := configProxy{}
	if err := unmarshal(&p); err != nil {
		return err
	}
	// validate modules
	enabled := make(map[string]struct{}, len(p.Enabled))
	for _, mod := range p.Enabled {
		switch mod {
		case "roster", "last_activity", "private", "vcard", "registration", "pep", "version", "blocking_command",
			"ping", "offline":
			break
		default:
			return fmt.Errorf("module.Config: unrecognized module: %s", mod)
		}
		enabled[mod] = struct{}{}
	}
	cfg.Enabled = enabled
	cfg.Roster = p.Roster
	cfg.Offline = p.Offline
	cfg.Version = p.Version
	cfg.Ping = p.Ping
	return nil
}
