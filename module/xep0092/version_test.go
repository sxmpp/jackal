/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package xep0092

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/sxmpp/jackal/router/host"

	c2srouter "github.com/sxmpp/jackal/c2s/router"
	"github.com/sxmpp/jackal/router"
	memorystorage "github.com/sxmpp/jackal/storage/memory"
	"github.com/sxmpp/jackal/stream"
	"github.com/sxmpp/jackal/version"
	"github.com/sxmpp/jackal/xmpp"
	"github.com/sxmpp/jackal/xmpp/jid"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
)

func TestXEP0092(t *testing.T) {
	r := setupTest()

	srvJID, _ := jid.New("", "jackal.im", "", true)
	j, _ := jid.New("sxmpp", "jackal.im", "balcony", true)

	stm := stream.NewMockC2S(uuid.New(), j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	cfg := Config{}
	x := New(&cfg, nil, r)
	defer func() { _ = x.Shutdown() }()

	// test MatchesIQ
	iq := xmpp.NewIQType(uuid.New(), xmpp.GetType)
	iq.SetFromJID(j)
	iq.SetToJID(j)

	qVer := xmpp.NewElementNamespace("query", versionNamespace)

	iq.AppendElement(xmpp.NewElementNamespace("query", "jabber:client"))
	require.False(t, x.MatchesIQ(iq))
	iq.ClearElements()
	iq.AppendElement(qVer)
	require.False(t, x.MatchesIQ(iq))
	iq.SetToJID(srvJID)
	require.True(t, x.MatchesIQ(iq))

	qVer.AppendElement(xmpp.NewElementName("version"))
	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	// get version
	qVer.ClearElements()
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	ver := elem.Elements().ChildNamespace("query", versionNamespace)
	require.Equal(t, "jackal", ver.Elements().Child("name").Text())
	require.Equal(t, version.ApplicationVersion.String(), ver.Elements().Child("version").Text())
	require.Nil(t, ver.Elements().Child("os"))

	// show OS
	cfg.ShowOS = true

	x = New(&cfg, nil, r)
	defer func() { _ = x.Shutdown() }()

	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	ver = elem.Elements().ChildNamespace("query", versionNamespace)
	require.Equal(t, osString, ver.Elements().Child("os").Text())
}

func setupTest() router.Router {
	hosts, _ := host.New([]host.Config{{Name: "jackal.im", Certificate: tls.Certificate{}}})
	r, _ := router.New(
		hosts,
		c2srouter.New(memorystorage.NewUser(), memorystorage.NewBlockList()),
		nil,
	)
	return r
}
