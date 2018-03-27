/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package model

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/ortuman/jackal/xml"
	"github.com/stretchr/testify/require"
)

func TestModelUser(t *testing.T) {
	var usr1, usr2 User

	usr1.Username = "ortuman"
	usr1.Password = "1234"

	buf := new(bytes.Buffer)
	usr1.ToGob(gob.NewEncoder(buf))
	usr2.FromGob(gob.NewDecoder(buf))
	require.Equal(t, usr1, usr2)
}

func TestModelRosterItem(t *testing.T) {
	var ri1, ri2 RosterItem

	ri1 = RosterItem{
		User:         "ortuman",
		Contact:      "noelia",
		Ask:          true,
		Subscription: "none",
		Groups:       []string{"friends", "family"},
	}
	buf := new(bytes.Buffer)
	ri1.ToGob(gob.NewEncoder(buf))
	ri2.FromGob(gob.NewDecoder(buf))
	require.Equal(t, ri1, ri2)
}

func TestModelRosterNotification(t *testing.T) {
	var rn1, rn2 RosterNotification

	rn1 = RosterNotification{
		User:     "ortuman",
		Contact:  "noelia",
		Elements: []xml.ElementNode{xml.NewElementNamespace("c", "http://jabber.org/protocol/caps")},
	}
	buf := new(bytes.Buffer)
	rn1.ToGob(gob.NewEncoder(buf))
	rn2.FromGob(gob.NewDecoder(buf))
	require.Equal(t, rn1, rn2)
}
