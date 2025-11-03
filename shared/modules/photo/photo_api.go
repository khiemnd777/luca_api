package photoApi

import (
	"context"

	batchUtil "github.com/khiemnd777/andy_api/shared/batch"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

func BatchGetPhotosByIDs(ctx context.Context, access string, photoIDs []int) ([]*generated.Photo, error) {
	return batchUtil.BatchApiGetByIDs[generated.Photo](ctx, "photo", access, photoIDs)
}
