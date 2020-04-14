/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package memorystorage

import (
	"context"
	"strings"

	"github.com/ortuman/jackal/model"
	capsmodel "github.com/ortuman/jackal/model/capabilities"
	"github.com/ortuman/jackal/model/serializer"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

type Presences struct {
	*memoryStorage
}

// NewPresences returns an instance of Presences in-memory storage.
func NewPresences() *Presences {
	return &Presences{memoryStorage: newStorage()}
}

func (m *Presences) UpsertPresence(_ context.Context, presence *xmpp.Presence, jid *jid.JID, allocationID string) (inserted bool, err error) {
	var ok bool
	k := presenceKey(jid, allocationID)
	if err := m.inWriteLock(func() error {
		_, ok = m.b[k]
		b, err := serializer.Serialize(presence)
		if err != nil {
			return err
		}
		m.b[k] = b
		return nil
	}); err != nil {
		return false, err
	}
	return !ok, nil
}

func (m *Presences) FetchPresence(_ context.Context, jid *jid.JID) (*model.ExtPresence, error) {
	var res *model.ExtPresence

	if err := m.inReadLock(func() error {
		for k, v := range m.b {
			if !strings.HasPrefix(k, "presences:"+jid.String()) {
				continue
			}
			extPresence, err := m.deserializePresence(v)
			if err != nil {
				return err
			}
			extPresence.AllocationID = allocationIDFromKey(k)
			res = extPresence
			return nil
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Presences) FetchPrioritaryPresence(ctx context.Context, jid *jid.JID) (*model.ExtPresence, error) {
	// TODO(ortuman): implement me!
	return nil, nil
}

func (m *Presences) FetchPresencesMatchingJID(ctx context.Context, j *jid.JID) ([]model.ExtPresence, error) {
	var usePrefix, useSuffix bool
	var res []model.ExtPresence

	if j.IsFullWithUser() {
		pCaps, err := m.FetchPresence(ctx, j)
		if err != nil {
			return nil, err
		}
		if pCaps == nil {
			return nil, nil
		}
		return []model.ExtPresence{*pCaps}, nil
	}
	usePrefix = j.IsBare()
	useSuffix = j.IsFullWithServer()

	if err := m.inReadLock(func() error {
		for k, v := range m.b {
			if !strings.HasPrefix(k, "presences:") {
				continue
			}
			ss := strings.Split(k, ":")

			kJID, _ := jid.NewWithString(ss[1], true)
			if usePrefix {
				if !j.MatchesWithOptions(kJID, jid.MatchesBare) {
					continue
				}
			} else if useSuffix {
				if !j.MatchesWithOptions(kJID, jid.MatchesDomain|jid.MatchesResource) {
					continue
				}
			} else if !j.MatchesWithOptions(kJID, jid.MatchesDomain) {
				continue
			}
			extPresence, err := m.deserializePresence(v)
			if err != nil {
				return err
			}
			extPresence.AllocationID = allocationIDFromKey(k)
			res = append(res, *extPresence)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Presences) DeletePresence(_ context.Context, jid *jid.JID) error {
	return m.inWriteLock(func() error {
		for k := range m.b {
			if strings.HasPrefix(k, "presences:"+jid.String()) {
				delete(m.b, k)
				return nil
			}
		}
		return nil
	})
}

func (m *Presences) DeleteAllocationPresences(_ context.Context, allocationID string) error {
	return m.inWriteLock(func() error {
		for k := range m.b {
			if strings.HasPrefix(k, "presences:") && strings.HasSuffix(k, ":"+allocationID) {
				delete(m.b, k)
			}
		}
		return nil
	})
}

func (m *Presences) FetchPresenceAllocationID(_ context.Context, jid *jid.JID) (string, error) {
	var allocID string
	if err := m.inReadLock(func() error {
		for k := range m.b {
			if strings.HasPrefix(k, "presences:"+jid.String()) {
				ss := strings.Split(k, ":")
				allocID = ss[2]
				return nil
			}
		}
		return nil
	}); err != nil {
		return "", err
	}
	return allocID, nil
}

func (m *Presences) FetchAllocationIDs(_ context.Context) ([]string, error) {
	allocationIDs := make(map[string]struct{})
	if err := m.inReadLock(func() error {
		for k := range m.b {
			ss := strings.Split(k, ":")
			allocationIDs[ss[2]] = struct{}{}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	var res []string
	for allocID := range allocationIDs {
		res = append(res, allocID)
	}
	return res, nil
}

func (m *Presences) UpsertCapabilities(_ context.Context, caps *capsmodel.Capabilities) error {
	return m.saveEntity(capabilitiesKey(caps.Node, caps.Ver), caps)
}

func (m *Presences) FetchCapabilities(_ context.Context, node, ver string) (*capsmodel.Capabilities, error) {
	var caps capsmodel.Capabilities

	ok, err := m.getEntity(capabilitiesKey(node, ver), &caps)
	switch err {
	case nil:
		if !ok {
			return nil, nil
		}
		return &caps, nil
	default:
		return nil, err
	}
}

func (m *Presences) deserializePresence(b []byte) (*model.ExtPresence, error) {
	var extPresence model.ExtPresence
	var presence xmpp.Presence

	if err := serializer.Deserialize(b, &presence); err != nil {
		return nil, err
	}
	extPresence.Presence = &presence
	if c := presence.Capabilities(); c != nil {
		if capsB := m.b[capabilitiesKey(c.Node, c.Ver)]; capsB != nil {
			var caps capsmodel.Capabilities
			if err := serializer.Deserialize(capsB, &caps); err != nil {
				return nil, err
			}
			extPresence.Caps = &caps
		}
	}
	return &extPresence, nil
}

func presenceKey(jid *jid.JID, allocationID string) string {
	return "presences:" + jid.String() + ":" + allocationID
}

func capabilitiesKey(node, ver string) string {
	return "capabilities:" + node + ":" + ver
}

func allocationIDFromKey(k string) string {
	ss := strings.Split(k, ":")
	if len(ss) != 3 {
		return ""
	}
	return ss[2]
}
