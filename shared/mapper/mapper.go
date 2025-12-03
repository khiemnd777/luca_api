package mapper

import (
	"reflect"
	"strings"
)

func Map[T any](item *T) T {
	if item == nil {
		var zero T
		return zero
	}
	return *item
}

func MapWith[T any](item *T, cb func(T) T) T {
	if item == nil {
		var zero T
		return zero
	}
	v := *item
	return cb(v)
}

func MapList[T any](items []*T) []T {
	if len(items) == 0 {
		if items == nil {
			return nil
		}
		return []T{}
	}
	out := make([]T, len(items))
	for i, it := range items {
		if it != nil {
			out[i] = *it
		}
	}
	return out
}

func MapAs[T any, R any](item T) R {
	var zeroR R
	src, ok := derefValue(reflect.ValueOf(item))
	if !ok {
		return zeroR
	}

	dstT := reflect.TypeOf(zeroR)
	dst, dstPtr := allocForType(dstT)

	// nếu R không phải struct, thử assign/convert thẳng (ít gặp)
	if dst.Kind() != reflect.Struct {
		if src.Type().AssignableTo(dst.Type()) {
			dst.Set(src)
			return castOut[R](dst, dstPtr)
		}
		if src.Type().ConvertibleTo(dst.Type()) {
			dst.Set(src.Convert(dst.Type()))
			return castOut[R](dst, dstPtr)
		}
		return zeroR
	}
	// nếu src không phải struct, cũng thử assign/convert
	if src.Kind() != reflect.Struct {
		if src.Type().AssignableTo(dst.Type()) {
			dst.Set(src)
			return castOut[R](dst, dstPtr)
		}
		if src.Type().ConvertibleTo(dst.Type()) {
			dst.Set(src.Convert(dst.Type()))
			return castOut[R](dst, dstPtr)
		}
		return zeroR
	}

	mapStruct(src, dst)
	return castOut[R](dst, dstPtr)
}

func MapListAs[T any, R any](items []T) []R {
	if len(items) == 0 {
		if items == nil {
			return nil
		}
		return []R{}
	}
	out := make([]R, len(items))
	for i, it := range items {
		out[i] = MapAs[T, R](it)
	}
	return out
}

func MapGet[T any, R any](item T, cb func(T) R) R {
	return cb(item)
}

func MapListGet[T any, R any](items []T, cb func(T) R) []R {
	if len(items) == 0 {
		if items == nil {
			return nil
		}
		return []R{}
	}
	out := make([]R, len(items))
	for i, it := range items {
		out[i] = cb(it)
	}
	return out
}

// ===== helpers =====
func derefValue(v reflect.Value) (reflect.Value, bool) {
	if !v.IsValid() {
		return reflect.Value{}, false
	}
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}, false
		}
		v = v.Elem()
	}
	return v, true
}
func allocForType(t reflect.Type) (reflect.Value, bool) {
	if t.Kind() == reflect.Pointer {
		elem := t.Elem()
		return reflect.New(elem).Elem(), true
	}
	return reflect.New(t).Elem(), false
}
func castOut[R any](v reflect.Value, dstWasPtr bool) R {
	var zero R
	t := reflect.TypeOf(zero)
	if dstWasPtr {
		p := reflect.New(v.Type())
		p.Elem().Set(v)
		return p.Convert(t).Interface().(R)
	}
	return v.Convert(t).Interface().(R)
}
func normalizeName(s string) string { return strings.ToLower(strings.ReplaceAll(s, "_", "")) }
func buildSrcIndex(src reflect.Value) map[string]int {
	idx := make(map[string]int, src.NumField())
	st := src.Type()
	for i := 0; i < src.NumField(); i++ {
		sf := st.Field(i)
		name := sf.Name
		if tag := sf.Tag.Get("map"); tag != "" {
			name = tag
		}
		idx[normalizeName(name)] = i
	}
	return idx
}
func isSimpleStruct(t reflect.Type) bool {
	return t.PkgPath() != ""
}
func isStructLike(v reflect.Value) bool {
	t := v.Type()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	return !isSimpleStruct(t)
}

func ptrOrValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return v.Elem()
	}
	return v
}
func assignBack(dst reflect.Value, src reflect.Value) {
	if dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst.Elem().Set(src)
	} else {
		dst.Set(src)
	}
}
func mapStruct(src, dst reflect.Value) {
	dstT := dst.Type()
	// index nguồn
	srcIdx := buildSrcIndex(src)
	for j := 0; j < dst.NumField(); j++ {
		df := dstT.Field(j)
		dv := dst.Field(j)
		if !dv.CanSet() {
			continue
		}
		srcName := df.Tag.Get("map")
		if srcName == "" {
			srcName = df.Name
		}
		if i, ok := srcIdx[normalizeName(srcName)]; ok {
			sv := src.Field(i)
			// struct lồng nhau: map nông
			if isStructLike(sv) && isStructLike(dv) {
				sve, ok := derefValue(sv)
				if !ok {
					continue
				}
				dve, _ := derefValue(ptrOrValue(dv))
				mapStruct(sve, dve)
				assignBack(dv, dve)
				continue
			}
			if sv.Type().AssignableTo(dv.Type()) {
				dv.Set(sv)
			} else if sv.Type().ConvertibleTo(dv.Type()) {
				dv.Set(sv.Convert(dv.Type()))
			}
		}
	}
}
