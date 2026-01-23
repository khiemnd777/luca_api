package engine

import (
	"encoding/json"
	"errors"
	"strconv"
)

func parseIntValue(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, errors.New("missing value")
	}
	var v int
	if json.Unmarshal(raw, &v) == nil {
		return v, nil
	}
	var f float64
	if json.Unmarshal(raw, &f) == nil {
		return int(f), nil
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return strconv.Atoi(s)
	}
	return 0, errors.New("invalid int value")
}

func parseIntList(raw json.RawMessage) ([]int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var out []int
	if json.Unmarshal(raw, &out) == nil {
		return out, nil
	}
	var any []any
	if json.Unmarshal(raw, &any) == nil {
		for _, v := range any {
			switch t := v.(type) {
			case float64:
				out = append(out, int(t))
			case string:
				i, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				out = append(out, i)
			}
		}
		return out, nil
	}
	return nil, errors.New("invalid int list")
}

func parseStringList(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var out []string
	if json.Unmarshal(raw, &out) == nil {
		return out, nil
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return []string{s}, nil
	}
	return nil, errors.New("invalid string list")
}

func containsInt(list []int, target int) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func anyInSet(orderIDs []int, allowed []int) bool {
	for _, id := range orderIDs {
		if containsInt(allowed, id) {
			return true
		}
	}
	return false
}

func anyInMap(ids []int, allowed map[int]struct{}) bool {
	for _, id := range ids {
		if _, ok := allowed[id]; ok {
			return true
		}
	}
	return false
}
