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
	builder.SetCustomFields(vr.Clean)
	return nil
}

func PrepareCustomFields[T CustomFieldsSetter[T]](
	ctx context.Context,
	validator *Manager,
	slugList []string,
	incoming map[string]any,
	builder T,
	isPatch bool,
) (map[string]any, error) {

	if len(slugList) == 0 {
		return nil, fmt.Errorf("no collection slugs provided")
	}

	merged := map[string]any{}
	allErrs := map[string]map[string]string{}

	for _, slug := range slugList {
		vr, err := validator.Validate(ctx, slug, incoming, isPatch)
		if err != nil {
			return nil, fmt.Errorf("validation error on %s: %w", slug, err)
		}

		if len(vr.Errs) > 0 {
			allErrs[slug] = vr.Errs
			continue
		}

		// Merge clean values (override theo thứ tự slugList)
		for k, v := range vr.Clean {
			merged[k] = v
		}
	}

	if len(allErrs) > 0 {
		return nil, fmt.Errorf("custom fields validation failed: %+v", allErrs)
	}

	builder.SetCustomFields(merged)

	return merged, nil
}
