/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package badgerdb

import (
	"github.com/dgraph-io/badger"
	"github.com/ortuman/jackal/xml"
)

func (b *Storage) InsertOrUpdatePrivateXML(privateXML []xml.XElement, namespace string, username string) error {
	r := xml.NewElementName("r")
	r.AppendElements(privateXML)
	return b.db.Update(func(tx *badger.Txn) error {
		return b.insertOrUpdate(r, b.privateStorageKey(username, namespace), tx)
	})
}

func (b *Storage) FetchPrivateXML(namespace string, username string) ([]xml.XElement, error) {
	var r xml.Element
	err := b.fetch(&r, b.privateStorageKey(username, namespace))
	switch err {
	case nil:
		return r.Elements().All(), nil
	case errBadgerDBEntityNotFound:
		return nil, nil
	default:
		return nil, err
	}
}

func (b *Storage) privateStorageKey(username, namespace string) []byte {
	return []byte("privateElements:" + username + ":" + namespace)
}