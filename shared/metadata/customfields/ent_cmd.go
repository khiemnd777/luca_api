package customfields

import (
	"context"
	"fmt"
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
	builder.SetCustomFields(customFields)
	return nil
}
