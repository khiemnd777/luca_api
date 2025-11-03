package userApi

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/shared/app"
	batchUtil "github.com/khiemnd777/andy_api/shared/batch"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

func BatchGetUsersByIDs(ctx context.Context, access string, userIDs []int) ([]*generated.User, error) {
	return batchUtil.BatchApiGetByIDs[generated.User](ctx, "user", access, userIDs)
}

func GetByID(ctx context.Context, access string, userID int) (*generated.User, error) {
	var result *generated.User

	if err := app.GetHttpClient().CallGet(ctx, "user", fmt.Sprintf("/api/user/%d", userID), access, "", &result); err != nil {
		return nil, err
	}

	return result, nil
}
