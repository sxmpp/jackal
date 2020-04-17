/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package pgsql

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/sxmpp/jackal/model"
	"github.com/sxmpp/jackal/util/pool"
	"github.com/sxmpp/jackal/xmpp"
	"github.com/sxmpp/jackal/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestInsertUser(t *testing.T) {
	from, _ := jid.NewWithString("sxmpp@jackal.im/Psi+", true)
	to, _ := jid.NewWithString("sxmpp@jackal.im", true)
	p := xmpp.NewPresence(from, to, xmpp.UnavailableType)

	user := model.User{Username: "sxmpp", Password: "1234", LastPresence: p}

	s, mock := newUserMock()
	mock.ExpectExec("INSERT INTO users (.+) ON CONFLICT (.+) DO UPDATE SET (.+)").
		WithArgs(user.Username, user.Password, user.LastPresence.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.UpsertUser(context.Background(), &user)
	require.Nil(t, err)
	require.Nil(t, mock.ExpectationsWereMet())

	s, mock = newUserMock()
	mock.ExpectExec("INSERT INTO users (.+) ON CONFLICT (.+) DO UPDATE SET (.+)").
		WithArgs(user.Username, user.Password, user.LastPresence.String()).
		WillReturnError(errMocked)

	err = s.UpsertUser(context.Background(), &user)
	require.Equal(t, errMocked, err)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestDeleteUser(t *testing.T) {
	s, mock := newUserMock()
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM offline_messages (.+)").
		WithArgs("sxmpp").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM roster_items (.+)").
		WithArgs("sxmpp").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM roster_versions (.+)").
		WithArgs("sxmpp").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM private_storage (.+)").
		WithArgs("sxmpp").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM vcards (.+)").
		WithArgs("sxmpp").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM users (.+)").
		WithArgs("sxmpp").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.DeleteUser(context.Background(), "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newUserMock()
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM offline_messages (.+)").
		WithArgs("sxmpp").WillReturnError(errMocked)
	mock.ExpectRollback()

	err = s.DeleteUser(context.Background(), "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMocked, err)
}

func TestFetchUser(t *testing.T) {
	from, _ := jid.NewWithString("sxmpp@jackal.im/Psi+", true)
	to, _ := jid.NewWithString("sxmpp@jackal.im", true)
	p := xmpp.NewPresence(from, to, xmpp.UnavailableType)

	var userColumns = []string{"username", "password", "last_presence", "last_presence_at"}

	s, mock := newUserMock()
	mock.ExpectQuery("SELECT (.+) FROM users (.+)").
		WithArgs("sxmpp").
		WillReturnRows(sqlmock.NewRows(userColumns))

	usr, _ := s.FetchUser(context.Background(), "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, usr)

	s, mock = newUserMock()
	mock.ExpectQuery("SELECT (.+) FROM users (.+)").
		WithArgs("sxmpp").
		WillReturnRows(sqlmock.NewRows(userColumns).AddRow("sxmpp", "1234", p.String(), time.Now()))
	_, err := s.FetchUser(context.Background(), "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newUserMock()
	mock.ExpectQuery("SELECT (.+) FROM users (.+)").
		WithArgs("sxmpp").WillReturnError(errMocked)
	_, err = s.FetchUser(context.Background(), "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMocked, err)
}

func TestUserExists(t *testing.T) {
	countColums := []string{"count"}

	s, mock := newUserMock()
	mock.ExpectQuery("SELECT COUNT(.+) FROM users (.+)").
		WithArgs("sxmpp").
		WillReturnRows(sqlmock.NewRows(countColums).AddRow(1))

	ok, err := s.UserExists(context.Background(), "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.True(t, ok)

	s, mock = newUserMock()
	mock.ExpectQuery("SELECT COUNT(.+) FROM users (.+)").
		WithArgs("romeo").
		WillReturnError(errMocked)
	_, err = s.UserExists(context.Background(), "romeo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMocked, err)
}

func newUserMock() (*pgSQLUser, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &pgSQLUser{
		pgSQLStorage: s,
		pool:         pool.NewBufferPool(),
	}, sqlMock
}
