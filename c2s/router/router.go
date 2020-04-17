/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package c2srouter

import (
	"context"
	"sort"
	"sync"

	"github.com/ortuman/jackal/cluster"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/model"
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
	// fetch user extended presences
	extPresences, err := r.presencesSt.FetchPresencesMatchingJID(ctx, stanza.ToJID().ToBareJID())
	if err != nil {
		return err
	}
	if len(extPresences) == 0 {
		return router.ErrNotAuthenticated
	}
	return r.route(ctx, stanza, extPresences)
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

func (r *c2sRouter) route(ctx context.Context, stanza xmpp.Stanza, extPresences []model.ExtPresence) error {
	toJID := stanza.ToJID()
	if toJID.IsFullWithUser() {
		return r.routeToResource(ctx, stanza, extPresences)
	}
	switch msg := stanza.(type) {
	case *xmpp.Message:
		routed, err := r.routeToPrioritaryResources(ctx, msg, extPresences)
		if err != nil {
			return err
		}
		if !routed {
			goto route2all
		}
		return nil
	}
route2all:
	return r.routeToAllResources(ctx, stanza, extPresences)
}

func (r *c2sRouter) routeToResource(ctx context.Context, stanza xmpp.Stanza, extPresences []model.ExtPresence) error {
	for _, extPresence := range extPresences {
		if stanza.ToJID().Resource() != extPresence.Presence.FromJID().Resource() {
			continue
		}
		return r.routeToAllocation(ctx, stanza, extPresence.AllocationID)
	}
	return router.ErrResourceNotFound
}

func (r *c2sRouter) routeToPrioritaryResources(ctx context.Context, stanza xmpp.Stanza, extPresences []model.ExtPresence) (routed bool, err error) {
	sort.Slice(extPresences, func(i, j int) bool {
		return extPresences[i].Presence.Priority() > extPresences[j].Presence.Priority()
	})
	highestPriority := extPresences[0].Presence.Priority()
	if highestPriority == 0 {
		return false, nil // no prioritary presence found
	}
	var prioritaryPresences []model.ExtPresence
	for _, extPresence := range extPresences {
		if extPresence.Presence.Priority() != highestPriority {
			break
		}
		prioritaryPresences = append(prioritaryPresences, extPresence)
	}
	// broacast to prioritary resources
	if err := r.routeToAllResources(ctx, stanza, prioritaryPresences); err != nil {
		return false, err
	}
	return true, nil
}

func (r *c2sRouter) routeToAllResources(ctx context.Context, stanza xmpp.Stanza, extPresences []model.ExtPresence) error {
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
	if r.clusterRouter == nil || r.cluster.IsLocalAllocationID(allocID) {
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
