/*
 * Copyright (c) 2019 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package storage

import (
	"fmt"

	memorystorage "github.com/sxmpp/jackal/storage/memory"
	"github.com/sxmpp/jackal/storage/mysql"
	"github.com/sxmpp/jackal/storage/pgsql"
	"github.com/sxmpp/jackal/storage/repository"
)

// New initializes configured storage type and returns associated container.
func New(config *Config) (repository.Container, error) {
	switch config.Type {
	case MySQL:
		return mysql.New(config.MySQL)
	case PostgreSQL:
		return pgsql.New(config.PostgreSQL)
	case Memory:
		return memorystorage.New()
	default:
		return nil, fmt.Errorf("storage: unrecognized storage type: %d", config.Type)
	}
}
