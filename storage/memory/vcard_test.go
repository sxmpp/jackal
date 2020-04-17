/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package memorystorage

import (
	"context"
	"testing"

	"github.com/sxmpp/jackal/xmpp"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_InsertVCard(t *testing.T) {
	vCard := xmpp.NewElementName("vCard")
	fn := xmpp.NewElementName("FN")
	fn.SetText("Miguel Ángel")
	vCard.AppendElement(fn)

	s := NewVCard()
	EnableMockedError()
	require.Equal(t, ErrMocked, s.UpsertVCard(context.Background(), vCard, "sxmpp"))
	DisableMockedError()
	require.Nil(t, s.UpsertVCard(context.Background(), vCard, "sxmpp"))
}

func TestMemoryStorage_FetchVCard(t *testing.T) {
	vCard := xmpp.NewElementName("vCard")
	fn := xmpp.NewElementName("FN")
	fn.SetText("Miguel Ángel")
	vCard.AppendElement(fn)

	s := NewVCard()
	_ = s.UpsertVCard(context.Background(), vCard, "sxmpp")

	EnableMockedError()
	_, err := s.FetchVCard(context.Background(), "sxmpp")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	elem, _ := s.FetchVCard(context.Background(), "sxmpp")
	require.NotNil(t, elem)
}
