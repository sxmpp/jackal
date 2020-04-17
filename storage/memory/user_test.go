/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package memorystorage

import (
	"context"
	"testing"

	"github.com/sxmpp/jackal/model"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_InsertUser(t *testing.T) {
	u := model.User{Username: "sxmpp", Password: "1234"}
	s := NewUser()
	EnableMockedError()
	err := s.UpsertUser(context.Background(), &u)
	require.Equal(t, ErrMocked, err)
	DisableMockedError()
	err = s.UpsertUser(context.Background(), &u)
	require.Nil(t, err)
}

func TestMemoryStorage_UserExists(t *testing.T) {
	s := NewUser()
	EnableMockedError()
	_, err := s.UserExists(context.Background(), "sxmpp")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()
	ok, err := s.UserExists(context.Background(), "sxmpp")
	require.Nil(t, err)
	require.False(t, ok)
}

func TestMemoryStorage_FetchUser(t *testing.T) {
	u := model.User{Username: "sxmpp", Password: "1234"}
	s := NewUser()
	_ = s.UpsertUser(context.Background(), &u)

	EnableMockedError()
	_, err := s.FetchUser(context.Background(), "sxmpp")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	usr, _ := s.FetchUser(context.Background(), "romeo")
	require.Nil(t, usr)

	usr, _ = s.FetchUser(context.Background(), "sxmpp")
	require.NotNil(t, usr)
}

func TestMemoryStorage_DeleteUser(t *testing.T) {
	u := model.User{Username: "sxmpp", Password: "1234"}
	s := NewUser()
	_ = s.UpsertUser(context.Background(), &u)

	EnableMockedError()
	require.Equal(t, ErrMocked, s.DeleteUser(context.Background(), "sxmpp"))
	DisableMockedError()
	require.Nil(t, s.DeleteUser(context.Background(), "sxmpp"))

	usr, _ := s.FetchUser(context.Background(), "sxmpp")
	require.Nil(t, usr)
}
