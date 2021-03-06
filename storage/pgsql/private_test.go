/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package pgsql

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/sxmpp/jackal/util/pool"
	"github.com/sxmpp/jackal/xmpp"
	"github.com/stretchr/testify/require"
)

func TestInsertPrivateXML(t *testing.T) {
	private := xmpp.NewElementNamespace("exodus", "exodus:ns")
	rawXML := private.String()

	s, mock := newPrivateMock()
	mock.ExpectExec("INSERT INTO private_storage (.+) ON CONFLICT (.+) DO UPDATE SET (.+)").
		WithArgs("sxmpp", "exodus:ns", rawXML, rawXML).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.UpsertPrivateXML(context.Background(), []xmpp.XElement{private}, "exodus:ns", "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newPrivateMock()
	mock.ExpectExec("INSERT INTO private_storage (.+) ON CONFLICT (.+) DO UPDATE SET (.+)").
		WithArgs("sxmpp", "exodus:ns", rawXML, rawXML).
		WillReturnError(errGeneric)

	err = s.UpsertPrivateXML(context.Background(), []xmpp.XElement{private}, "exodus:ns", "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errGeneric, err)
}

func TestFetchPrivateXML(t *testing.T) {
	var privateColumns = []string{"data"}

	s, mock := newPrivateMock()
	mock.ExpectQuery("SELECT (.+) FROM private_storage (.+)").
		WithArgs("sxmpp", "exodus:ns").
		WillReturnRows(sqlmock.NewRows(privateColumns).AddRow("<exodus xmlns='exodus:ns'><stuff/></exodus>"))

	elems, err := s.FetchPrivateXML(context.Background(), "exodus:ns", "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Equal(t, 1, len(elems))

	s, mock = newPrivateMock()
	mock.ExpectQuery("SELECT (.+) FROM private_storage (.+)").
		WithArgs("sxmpp", "exodus:ns").
		WillReturnRows(sqlmock.NewRows(privateColumns).AddRow("<exodus xmlns='exodus:ns'><stuff/>"))

	elems, err = s.FetchPrivateXML(context.Background(), "exodus:ns", "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.NotNil(t, err)
	require.Equal(t, 0, len(elems))

	s, mock = newPrivateMock()
	mock.ExpectQuery("SELECT (.+) FROM private_storage (.+)").
		WithArgs("sxmpp", "exodus:ns").
		WillReturnRows(sqlmock.NewRows(privateColumns).AddRow(""))

	elems, err = s.FetchPrivateXML(context.Background(), "exodus:ns", "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Equal(t, 0, len(elems))

	s, mock = newPrivateMock()
	mock.ExpectQuery("SELECT (.+) FROM private_storage (.+)").
		WithArgs("sxmpp", "exodus:ns").
		WillReturnRows(sqlmock.NewRows(privateColumns))

	elems, err = s.FetchPrivateXML(context.Background(), "exodus:ns", "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Equal(t, 0, len(elems))

	s, mock = newPrivateMock()
	mock.ExpectQuery("SELECT (.+) FROM private_storage (.+)").
		WithArgs("sxmpp", "exodus:ns").
		WillReturnError(errGeneric)

	elems, err = s.FetchPrivateXML(context.Background(), "exodus:ns", "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errGeneric, err)
	require.Equal(t, 0, len(elems))
}

func newPrivateMock() (*pgSQLPrivate, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &pgSQLPrivate{
		pgSQLStorage: s,
		pool:         pool.NewBufferPool(),
	}, sqlMock
}
