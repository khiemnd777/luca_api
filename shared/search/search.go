package searchutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
)

func BuildKeywords(
	ctx context.Context,
	cfMgr *customfields.Manager,
	slug string,
	core []any,
	customFields map[string]any,
) (string, error) {

	parts := make([]string, 0, 16)

	for _, v := range core {
		switch val := v.(type) {

		case string:
			if val != "" {
				parts = append(parts, val)
			}

		case *string:
			if val != nil && *val != "" {
				parts = append(parts, *val)
			}

		case []string:
			for _, s := range val {
				if s != "" {
					parts = append(parts, s)
				}
			}

		case []*string:
			for _, s := range val {
				if s != nil && *s != "" {
					parts = append(parts, *s)
				}
			}

		case []any:
			for _, x := range val {
				s := fmt.Sprint(x)
				if s != "" && s != "0" && s != "false" {
					parts = append(parts, s)
				}
			}

		default:
			s := fmt.Sprint(val)
			if s != "" && s != "0" && s != "false" {
				parts = append(parts, s)
			}
		}
	}

	if customFields != nil {
		cfParts, err := cfMgr.GetSearchFieldValues(ctx, slug, customFields)
		if err != nil {
			return "", err
		}
		parts = append(parts, cfParts...)
	}

	if len(parts) == 0 {
		return "", nil
	}

	return strings.Join(parts, "|"), nil
}
