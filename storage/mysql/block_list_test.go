/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package mysql

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/sxmpp/jackal/model"
	"github.com/stretchr/testify/require"
)

func TestMySQLStorageInsertBlockListItems(t *testing.T) {
	s, mock := newBlockListMock()
	mock.ExpectExec("INSERT IGNORE INTO blocklist_items (.+)").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "sxmpp", JID: "noelia@jackal.im"})
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newBlockListMock()
	mock.ExpectExec("INSERT IGNORE INTO blocklist_items (.+)").WillReturnError(errMySQLStorage)

	err = s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "sxmpp", JID: "noelia@jackal.im"})
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLFetchBlockListItems(t *testing.T) {
	var blockListColumns = []string{"username", "jid"}
	s, mock := newBlockListMock()
	mock.ExpectQuery("SELECT (.+) FROM blocklist_items (.+)").
		WithArgs("sxmpp").
		WillReturnRows(sqlmock.NewRows(blockListColumns).AddRow("sxmpp", "noelia@jackal.im"))

	_, err := s.FetchBlockListItems(context.Background(), "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newBlockListMock()
	mock.ExpectQuery("SELECT (.+) FROM blocklist_items (.+)").
		WithArgs("sxmpp").
		WillReturnError(errMySQLStorage)

	_, err = s.FetchBlockListItems(context.Background(), "sxmpp")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorageDeleteBlockListItems(t *testing.T) {
	s, mock := newBlockListMock()
	mock.ExpectExec("DELETE FROM blocklist_items (.+)").
		WithArgs("sxmpp").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s, mock = newBlockListMock()
	mock.ExpectExec("DELETE FROM blocklist_items (.+)").
		WithArgs("sxmpp", "noelia@jackal.im").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.DeleteBlockListItem(context.Background(), &model.BlockListItem{Username: "sxmpp", JID: "noelia@jackal.im"})
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newBlockListMock()
	mock.ExpectExec("DELETE FROM blocklist_items (.+)").
		WillReturnError(errMySQLStorage)

	err = s.DeleteBlockListItem(context.Background(), &model.BlockListItem{Username: "sxmpp", JID: "noelia@jackal.im"})
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func newBlockListMock() (*mySQLBlockList, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &mySQLBlockList{
		mySQLStorage: s,
	}, sqlMock
}
