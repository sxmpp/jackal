/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package c2srouter

import (
	"context"
	"sync"

	"github.com/ortuman/jackal/cluster"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/storage"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

type c2sRouter struct {
	userSt        storage.User
	blockListSt   storage.BlockList
	presencesSt   storage.Presences
	cluster       *cluster.Cluster
	localRouter   *localRouter
	clusterRouter *clusterRouter
}

func New(userSt storage.User, blockListSt storage.BlockList, presencesSt storage.Presences, cluster *cluster.Cluster) (router.C2SRouter, error) {
	r := &c2sRouter{
		userSt:      userSt,
		blockListSt: blockListSt,
		presencesSt: presencesSt,
		localRouter: newLocalRouter(),
	}
	if cluster != nil {
		clusterRouter, err := newClusterRouter(cluster, presencesSt)
		if err != nil {
			return nil, err
		}
		r.cluster = cluster
		r.clusterRouter = clusterRouter

		// register local router as cluster stanza handler
		cluster.RegisterStanzaHandler(r.localRouter.route)
	}
	return r, nil
}

func (r *c2sRouter) Route(ctx context.Context, stanza xmpp.Stanza, validations router.C2SRoutingValidations) error {
	fromJID := stanza.FromJID()
	toJID := stanza.ToJID()

	// apply validations
	username := stanza.ToJID().Node()
	if (validations & router.UserExistence) > 0 {
		exists, err := r.userSt.UserExists(ctx, username) // user exists?
		if err != nil {
			return err
		}
		if !exists {
			return router.ErrNotExistingAccount
		}
	}
	if (validations & router.BlockedDestinationJID) > 0 {
		if r.isBlockedJID(ctx, toJID, fromJID.Node()) { // check whether destination JID is blocked
			return router.ErrBlockedJID
		}
	}
	return r.route(ctx, stanza)
}

func (r *c2sRouter) Bind(stm stream.C2S) {
	r.localRouter.bind(stm)

	log.Infof("bound c2s stream... (%s/%s)", stm.Username(), stm.Resource())
}

func (r *c2sRouter) Unbind(user, resource string) {
	r.localRouter.unbind(user, resource)

	log.Infof("unbound c2s stream... (%s/%s)", user, resource)
}

func (r *c2sRouter) Stream(username, resource string) stream.C2S {
	return r.localRouter.stream(username, resource)
}

func (r *c2sRouter) Streams(username string) []stream.C2S {
	return r.localRouter.streams(username)
}

func (r *c2sRouter) route(ctx context.Context, stanza xmpp.Stanza) error {
	toJID := stanza.ToJID()
	if toJID.IsFullWithUser() {
		return r.routeToResource(ctx, stanza)
	}
	switch msg := stanza.(type) {
	case *xmpp.Message:
		routed, err := r.routeMessageToPrioritaryResource(ctx, msg)
		if err != nil {
			return err
		}
		if !routed {
			goto route2all
		}
		return nil
	}
route2all:
	return r.routeToAllResources(ctx, stanza)
}

func (r *c2sRouter) routeToResource(ctx context.Context, stanza xmpp.Stanza) error {
	err := r.localRouter.route(ctx, stanza) // first, try to route locally
	switch err {
	case router.ErrResourceNotFound:
		if r.cluster == nil {
			return err
		}
		break
	default:
		return err
	}
	allocID, err := r.presencesSt.FetchPresenceAllocationID(ctx, stanza.ToJID())
	if err != nil {
		return err
	}
	if len(allocID) == 0 {
		return router.ErrResourceNotFound
	}
	return r.clusterRouter.route(ctx, stanza, allocID)
}

func (r *c2sRouter) routeMessageToPrioritaryResource(ctx context.Context, msg *xmpp.Message) (routed bool, err error) {
	extPresence, err := r.presencesSt.FetchPrioritaryPresence(ctx, msg.ToJID())
	if err != nil {
		return false, err
	}
	if extPresence == nil { // no prioritary presence found
		return false, nil
	}
	// rewrite message stanza pointing to prioritary resource
	rwMessage, err := xmpp.NewMessageFromElement(msg, msg.FromJID(), extPresence.Presence.FromJID())
	if err != nil {
		return false, err
	}
	if err := r.routeToAllocation(ctx, rwMessage, extPresence.AllocationID); err != nil {
		return false, err
	}
	return true, nil
}

func (r *c2sRouter) routeToAllResources(ctx context.Context, stanza xmpp.Stanza) error {
	extPresences, err := r.presencesSt.FetchPresencesMatchingJID(ctx, stanza.ToJID().ToBareJID())
	if err != nil {
		return err
	}
	allocationIDs := make(map[string]struct{})
	for _, extPresence := range extPresences {
		allocationIDs[extPresence.AllocationID] = struct{}{}
	}
	errCh := make(chan error, len(allocationIDs))

	var wg sync.WaitGroup
	for allocID := range allocationIDs {
		wg.Add(1)

		go func(allocationID string) {
			defer wg.Done()
			if err := r.routeToAllocation(ctx, stanza, allocationID); err != nil {
				errCh <- err
			}
		}(allocID)
	}
	go func() {
		wg.Wait()
		errCh <- nil
	}()
	return <-errCh
}

func (r *c2sRouter) routeToAllocation(ctx context.Context, stanza xmpp.Stanza, allocID string) error {
	if r.cluster == nil || r.cluster.IsLocalAllocationID(allocID) {
		return r.localRouter.route(ctx, stanza)
	}
	return r.clusterRouter.route(ctx, stanza, allocID)
}

func (r *c2sRouter) isBlockedJID(ctx context.Context, j *jid.JID, username string) bool {
	blockList, err := r.blockListSt.FetchBlockListItems(ctx, username)
	if err != nil {
		log.Error(err)
		return false
	}
	if len(blockList) == 0 {
		return false
	}
	blockListJIDs := make([]jid.JID, len(blockList))
	for i, listItem := range blockList {
		j, _ := jid.NewWithString(listItem.JID, true)
		blockListJIDs[i] = *j
	}
	for _, blockedJID := range blockListJIDs {
		if blockedJID.Matches(j) {
			return true
		}
	}
	return false
}
