/*
 * Copyright (c) 2019 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package pgsql

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	pubsubmodel "github.com/sxmpp/jackal/model/pubsub"
	"github.com/sxmpp/jackal/xmpp"
	"github.com/stretchr/testify/require"
)

func TestPgSQLFetchPubSubHosts(t *testing.T) {
	s, mock := newPubSubMock()
	rows := sqlmock.NewRows([]string{"host"})
	rows.AddRow("sxmpp@jackal.im")
	rows.AddRow("noelia@jackal.im")

	mock.ExpectQuery("SELECT DISTINCT\\(host\\) FROM pubsub_nodes").
		WillReturnRows(rows)

	hosts, err := s.FetchHosts(context.Background())
	require.Nil(t, err)
	require.NotNil(t, hosts)
	require.Equal(t, "sxmpp@jackal.im", hosts[0])
	require.Equal(t, "noelia@jackal.im", hosts[1])

	s, mock = newPubSubMock()
	mock.ExpectQuery("SELECT DISTINCT\\(host\\) FROM pubsub_nodes").
		WillReturnError(errGeneric)

	hosts, err = s.FetchHosts(context.Background())
	require.Nil(t, hosts)
	require.NotNil(t, err)
	require.Equal(t, errGeneric, err)
}

func TestPgSQLUpsertPubSubNode(t *testing.T) {
	s, mock := newPubSubMock()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO pubsub_nodes (.+) ON CONFLICT (.+) DO NOTHING").
		WithArgs("host", "name").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("SELECT id FROM pubsub_nodes WHERE (.+)").
		WithArgs("host", "name").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	mock.ExpectExec("DELETE FROM pubsub_node_options WHERE (.+)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	opts := pubsubmodel.Options{}

	optMap, _ := opts.Map()
	for i := 0; i < len(optMap); i++ {
		mock.ExpectExec("INSERT INTO pubsub_node_options (.+)").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	node := pubsubmodel.Node{Host: "host", Name: "name", Options: opts}
	err := s.UpsertNode(context.Background(), &node)

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnError(errGeneric)

	_, err = s.FetchNode(context.Background(), "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())

	require.NotNil(t, err)
	require.Equal(t, errGeneric, err)
}

func TestPgSQLFetchPubSubNode(t *testing.T) {
	var cols = []string{"name", "value"}

	s, mock := newPubSubMock()
	rows := sqlmock.NewRows(cols)
	rows.AddRow("pubsub#access_model", "presence")
	rows.AddRow("pubsub#publish_model", "publishers")
	rows.AddRow("pubsub#send_last_published_item", "on_sub_and_presence")

	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnRows(rows)

	node, err := s.FetchNode(context.Background(), "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)
	require.NotNil(t, node)
	require.Equal(t, node.Options.AccessModel, pubsubmodel.Presence)
	require.Equal(t, node.Options.SendLastPublishedItem, pubsubmodel.OnSubAndPresence)
}

func TestPgSQLFetchPubSubNodes(t *testing.T) {
	s, mock := newPubSubMock()
	rows := sqlmock.NewRows([]string{"name"})
	rows.AddRow("princely_musings_1")
	rows.AddRow("princely_musings_2")

	mock.ExpectQuery("SELECT name FROM pubsub_nodes WHERE host = (.+)").
		WithArgs("sxmpp@jackal.im").
		WillReturnRows(rows)

	var cols = []string{"name", "value"}

	rows = sqlmock.NewRows(cols)
	rows.AddRow("pubsub#access_model", "presence")
	rows.AddRow("pubsub#publish_model", "publishers")
	rows.AddRow("pubsub#send_last_published_item", "on_sub_and_presence")

	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings_1").
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings_2").
		WillReturnRows(rows)

	nodes, err := s.FetchNodes(context.Background(), "sxmpp@jackal.im")

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)
	require.NotNil(t, nodes)
	require.Len(t, nodes, 2)
	require.Equal(t, "princely_musings_1", nodes[0].Name)
	require.Equal(t, "princely_musings_2", nodes[1].Name)
}

func TestPgSQLFetchPubSubSubscribedNodes(t *testing.T) {
	s, mock := newPubSubMock()
	rows := sqlmock.NewRows([]string{"host", "name"})
	rows.AddRow("sxmpp@jackal.im", "princely_musings_1")
	rows.AddRow("sxmpp@jackal.im", "princely_musings_2")

	mock.ExpectQuery("SELECT host, name FROM pubsub_nodes WHERE id IN (.+)").
		WithArgs("sxmpp@jackal.im", pubsubmodel.Subscribed).
		WillReturnRows(rows)

	var cols = []string{"name", "value"}

	rows = sqlmock.NewRows(cols)
	rows.AddRow("pubsub#access_model", "presence")
	rows.AddRow("pubsub#publish_model", "publishers")
	rows.AddRow("pubsub#send_last_published_item", "on_sub_and_presence")

	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings_1").
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings_2").
		WillReturnRows(rows)

	nodes, err := s.FetchSubscribedNodes(context.Background(), "sxmpp@jackal.im")

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)
	require.NotNil(t, nodes)
	require.Len(t, nodes, 2)
	require.Equal(t, "princely_musings_1", nodes[0].Name)
	require.Equal(t, "princely_musings_2", nodes[1].Name)
}

func TestPgSQLDeletePubSubNode(t *testing.T) {
	s, mock := newPubSubMock()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM pubsub_nodes WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	mock.ExpectExec("DELETE FROM pubsub_nodes WHERE (.+)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM pubsub_node_options WHERE (.+)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM pubsub_items WHERE (.+)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM pubsub_affiliations WHERE (.+)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM pubsub_subscriptions WHERE (.+)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.DeleteNode(context.Background(), "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestPgSQLUpsertPubSubNodeItem(t *testing.T) {
	payload := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)

	s, mock := newPubSubMock()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM pubsub_nodes WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	mock.ExpectExec("INSERT INTO pubsub_items (.+) ON CONFLICT (.+) DO UPDATE SET payload = (.+), publisher = (.+)").
		WithArgs("1", "abc1234", payload.String(), "sxmpp@jackal.im", payload.String(), "sxmpp@jackal.im").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM pubsub_items WHERE item_id IN (.+)").
		WithArgs("1", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.UpsertNodeItem(context.Background(), &pubsubmodel.Item{
		ID:        "abc1234",
		Publisher: "sxmpp@jackal.im",
		Payload:   payload,
	}, "sxmpp@jackal.im", "princely_musings", 1)

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)
}

func TestPgSQLFetchPubSubNodeItems(t *testing.T) {
	s, mock := newPubSubMock()
	rows := sqlmock.NewRows([]string{"item_id", "publisher", "payload"})
	rows.AddRow("1234", "sxmpp@jackal.im", "<message/>")
	rows.AddRow("5678", "noelia@jackal.im", "<iq type='get'/>")

	mock.ExpectQuery("SELECT item_id, publisher, payload FROM pubsub_items WHERE node_id = (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnRows(rows)

	items, err := s.FetchNodeItems(context.Background(), "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)
	require.Equal(t, 2, len(items))
	require.Equal(t, "1234", items[0].ID)
	require.Equal(t, "5678", items[1].ID)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectQuery("SELECT item_id, publisher, payload FROM pubsub_items WHERE node_id = (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnError(errGeneric)

	_, err = s.FetchNodeItems(context.Background(), "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())

	require.NotNil(t, err)
	require.Equal(t, errGeneric, err)
}

func TestPgSQLFetchPubSubNodeItemsWithID(t *testing.T) {
	s, mock := newPubSubMock()
	rows := sqlmock.NewRows([]string{"item_id", "publisher", "payload"})
	rows.AddRow("1234", "sxmpp@jackal.im", "<message/>")
	rows.AddRow("5678", "noelia@jackal.im", "<iq type='get'/>")

	identifiers := []string{"1234", "5678"}

	mock.ExpectQuery("SELECT item_id, publisher, payload FROM pubsub_items WHERE (.+ IN (.+)) ORDER BY created_at").
		WithArgs("sxmpp@jackal.im", "princely_musings", "1234", "5678").
		WillReturnRows(rows)

	items, err := s.FetchNodeItemsWithIDs(context.Background(), "sxmpp@jackal.im", "princely_musings", identifiers)

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)
	require.Equal(t, 2, len(items))
	require.Equal(t, "1234", items[0].ID)
	require.Equal(t, "5678", items[1].ID)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectQuery("SELECT item_id, publisher, payload FROM pubsub_items WHERE (.+ IN (.+)) ORDER BY created_at").
		WithArgs("sxmpp@jackal.im", "princely_musings", "1234", "5678").
		WillReturnError(errGeneric)

	_, err = s.FetchNodeItemsWithIDs(context.Background(), "sxmpp@jackal.im", "princely_musings", identifiers)

	require.Nil(t, mock.ExpectationsWereMet())

	require.NotNil(t, err)
	require.Equal(t, errGeneric, err)
}

func TestPgSQLUpsertPubSubNodeAffiliation(t *testing.T) {
	s, mock := newPubSubMock()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM pubsub_nodes WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	mock.ExpectExec("INSERT INTO pubsub_affiliations (.+) VALUES (.+) ON CONFLICT (.+) DO UPDATE SET affiliation = (.+)").
		WithArgs("1", "sxmpp@jackal.im", "owner", "owner").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "sxmpp@jackal.im",
		Affiliation: "owner",
	}, "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)
}

func TestPgSQLFetchPubSubNodeAffiliations(t *testing.T) {
	s, mock := newPubSubMock()
	rows := sqlmock.NewRows([]string{"jid", "affiliation"})
	rows.AddRow("sxmpp@jackal.im", "owner")
	rows.AddRow("noelia@jackal.im", "publisher")

	mock.ExpectQuery("SELECT jid, affiliation FROM pubsub_affiliations WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnRows(rows)

	affiliations, err := s.FetchNodeAffiliations(context.Background(), "sxmpp@jackal.im", "princely_musings")
	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)
	require.Equal(t, 2, len(affiliations))

	// error case
	mock.ExpectQuery("SELECT jid, affiliation FROM pubsub_affiliations WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnError(errGeneric)

	affiliations, err = s.FetchNodeAffiliations(context.Background(), "sxmpp@jackal.im", "princely_musings")
	require.Nil(t, mock.ExpectationsWereMet())

	require.NotNil(t, err)
	require.Equal(t, errGeneric, err)
}

func TestPgSQLDeletePubSubNodeAffiliation(t *testing.T) {
	s, mock := newPubSubMock()

	mock.ExpectExec("DELETE FROM pubsub_affiliations WHERE (.+)").
		WithArgs("noeliac@jackal.im", "sxmpp@jackal.im", "princely_musings").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.DeleteNodeAffiliation(context.Background(), "noeliac@jackal.im", "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectExec("DELETE FROM pubsub_affiliations WHERE (.+)").
		WithArgs("noeliac@jackal.im", "sxmpp@jackal.im", "princely_musings").
		WillReturnError(errGeneric)

	err = s.DeleteNodeAffiliation(context.Background(), "noeliac@jackal.im", "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())

	require.NotNil(t, err)
	require.Equal(t, errGeneric, err)
}

func TestPgSQLUpsertPubSubNodeSubscription(t *testing.T) {
	s, mock := newPubSubMock()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM pubsub_nodes WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	mock.ExpectExec("INSERT INTO pubsub_subscriptions (.+) VALUES (.+) ON CONFLICT (.+) DO UPDATE SET (.+)").
		WithArgs("1", "1234", "sxmpp@jackal.im", "subscribed", "1234", "subscribed").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.UpsertNodeSubscription(context.Background(), &pubsubmodel.Subscription{
		SubID:        "1234",
		JID:          "sxmpp@jackal.im",
		Subscription: "subscribed",
	}, "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)
}

func TestPgSQLFetchPubSubNodeSubscriptions(t *testing.T) {
	s, mock := newPubSubMock()
	rows := sqlmock.NewRows([]string{"subid", "jid", "subscription"})
	rows.AddRow("1234", "sxmpp@jackal.im", "subscribed")
	rows.AddRow("5678", "noelia@jackal.im", "unsubscribed")

	mock.ExpectQuery("SELECT subid, jid, subscription FROM pubsub_subscriptions WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnRows(rows)

	subscriptions, err := s.FetchNodeSubscriptions(context.Background(), "sxmpp@jackal.im", "princely_musings")
	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)
	require.Equal(t, 2, len(subscriptions))

	// error case
	mock.ExpectQuery("SELECT subid, jid, subscription FROM pubsub_subscriptions WHERE (.+)").
		WithArgs("sxmpp@jackal.im", "princely_musings").
		WillReturnError(errGeneric)

	subscriptions, err = s.FetchNodeSubscriptions(context.Background(), "sxmpp@jackal.im", "princely_musings")
	require.Nil(t, mock.ExpectationsWereMet())

	require.NotNil(t, err)
	require.Equal(t, errGeneric, err)
}

func TestPgSQLDeletePubSubNodeSubscription(t *testing.T) {
	s, mock := newPubSubMock()

	mock.ExpectExec("DELETE FROM pubsub_subscriptions WHERE (.+)").
		WithArgs("noeliac@jackal.im", "sxmpp@jackal.im", "princely_musings").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.DeleteNodeSubscription(context.Background(), "noeliac@jackal.im", "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())

	require.Nil(t, err)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectExec("DELETE FROM pubsub_subscriptions WHERE (.+)").
		WithArgs("noeliac@jackal.im", "sxmpp@jackal.im", "princely_musings").
		WillReturnError(errGeneric)

	err = s.DeleteNodeSubscription(context.Background(), "noeliac@jackal.im", "sxmpp@jackal.im", "princely_musings")

	require.Nil(t, mock.ExpectationsWereMet())

	require.NotNil(t, err)
	require.Equal(t, errGeneric, err)
}

func newPubSubMock() (*pgSQLPubSub, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &pgSQLPubSub{
		pgSQLStorage: s,
	}, sqlMock
}
