/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package rostermodel

import (
	"bytes"
	"testing"

	"github.com/sxmpp/jackal/xmpp"
	"github.com/sxmpp/jackal/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestModelRosterNotification(t *testing.T) {
	var rn1, rn2 Notification

	j1, _ := jid.NewWithString("sxmpp@jackal.im", true)
	j2, _ := jid.NewWithString("noelia@jackal.im", true)

	rn1 = Notification{
		Contact:  "noelia",
		JID:      "sxmpp@jackal.im",
		Presence: xmpp.NewPresence(j1, j2, xmpp.AvailableType),
	}
	buf := new(bytes.Buffer)
	require.Nil(t, rn1.ToBytes(buf))
	require.Nil(t, rn2.FromBytes(buf))
	require.Equal(t, "sxmpp@jackal.im", rn2.JID)
	require.Equal(t, "noelia", rn2.Contact)
	require.NotNil(t, rn1.Presence)
	require.NotNil(t, rn2.Presence)
	require.Equal(t, rn1.Presence.String(), rn2.Presence.String())
}
