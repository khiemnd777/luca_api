package customfields

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/shared/logger"
)

type CustomFieldsSetter[T any] interface {
	SetCustomFields(map[string]any) T
}

func SetCustomFields[T CustomFieldsSetter[T]](
	ctx context.Context,
	validator *Manager,
	collectionSlug string,
	customFields map[string]any,
	builder T,
	isPatch bool,
) error {
	vr, err := validator.Validate(ctx, collectionSlug, customFields, isPatch)
	if err != nil {
		return err
	}
	if len(vr.Errs) > 0 {
		return fmt.Errorf("validation errors: %v", vr.Errs)
	}
	logger.Debug(fmt.Sprintf("[STAFF] VR %v", vr.Clean))
	builder.SetCustomFields(vr.Clean)
	return nil
}
