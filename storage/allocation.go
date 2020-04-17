package storage

import "context"

type Allocation interface {
	RegisterAllocation(ctx context.Context, allocationID string) error
	UnregisterAllocation(ctx context.Context, allocationID string) error

	FetchAllocations(ctx context.Context) (allocationIDs []string, err error)
}
