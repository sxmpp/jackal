package storage

import "github.com/ortuman/jackal/model"

type Resources interface {
	UpsertResource(resource *model.Resource, allocationID string) error

	FetchResources(username, domain string) ([]model.Resource, error)
}
