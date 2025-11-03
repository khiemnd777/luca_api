package assigner

import (
	"context"
)

func AssignEntities[T any, E any](
	ctx context.Context,
	items []T,
	extractID func(item T) int,
	extractEntityID func(*E) int,
	batchGetEntities func(ctx context.Context, ids []int) ([]*E, error),
	assignEntity func(item *T, entity *E),
) error {

	entityMap, err := fetchAndMapEntities(
		ctx,
		items,
		extractID,
		extractEntityID,
		batchGetEntities,
	)

	if err != nil {
		return err
	}

	for i := range items {
		id := extractID(items[i])
		if entity, ok := entityMap[id]; ok {
			assignEntity(&items[i], entity)
		}
	}

	return nil
}

func AssignEntitiesPtr[T any, E any](
	ctx context.Context,
	items []*T,
	extractID func(item *T) int,
	extractEntityID func(*E) int,
	batchGetEntities func(ctx context.Context, ids []int) ([]*E, error),
	assignEntity func(item *T, entity *E),
) error {
	entityMap, err := fetchAndMapEntities(
		ctx,
		items,
		extractID,
		extractEntityID,
		batchGetEntities,
	)

	if err != nil {
		return err
	}

	for _, item := range items {

		id := extractID(item)

		if entity, ok := entityMap[id]; ok {
			assignEntity(item, entity)
		}
	}

	return nil
}

func fetchAndMapEntities[T any, E any](
	ctx context.Context,
	items []T,
	extractID func(T) int,
	extractEntityID func(*E) int,
	batchGetEntities func(context.Context, []int) ([]*E, error),
) (map[int]*E, error) {
	if len(items) == 0 {
		return nil, nil
	}

	ids := extractIDs(items, extractID)

	entities, err := batchGetEntities(ctx, ids)
	if err != nil {
		return nil, err
	}

	entityMap := make(map[int]*E, len(entities))
	for _, e := range entities {
		entityMap[extractEntityID(e)] = e
	}

	return entityMap, nil
}

func extractIDs[T any](items []T, getID func(item T) int) []int {
	idSet := make(map[int]struct{})
	for _, item := range items {
		idSet[getID(item)] = struct{}{}
	}

	result := make([]int, 0, len(idSet))
	for id := range idSet {
		result = append(result, id)
	}

	return result
}
