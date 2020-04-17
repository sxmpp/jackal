package model

import "github.com/ortuman/jackal/xmpp/jid"

type Resource struct {
	AllocationID string
	JID          *jid.JID
	Priority     int8
}
