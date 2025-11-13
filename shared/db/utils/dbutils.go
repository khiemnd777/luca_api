package dbutils

import (
	"database/sql/driver"
	"strings"
)

type PqStringArray []string

func (a PqStringArray) Value() (driver.Value, error) { return "{" + strings.Join(a, ",") + "}", nil }
