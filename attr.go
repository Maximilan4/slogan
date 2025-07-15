package slogan

import (
	"fmt"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"time"
)

// Pointer - helper func for slog to write pointer T values
func Pointer[T any](key string, value *T) slog.Attr {
	if value == nil {
		return slog.String(key, "<nil>")
	}

	return Any(key, *value)
}

// Map - adds to log all k-v pairs as attributes, recursive
func Map[M ~map[K]V, K comparable, V any](key string, m M) slog.Attr {
	if m == nil {
		return slog.String(key, "<nil>")
	}

	return slog.Any(key, MapValue(m).Group())
}

// MapValue - prepares map group value
func MapValue[M ~map[K]V, K comparable, V any](m M) slog.Value {
	if m == nil {
		return slog.StringValue("<nil>")
	}

	if len(m) == 0 {
		return slog.StringValue("{}")
	}

	attrs := make([]slog.Attr, 0, len(m))

	var (
		k K
		v V
	)
	for k, v = range maps.All(m) {
		attrs = append(attrs, Any(fmt.Sprintf("%v", k), v))
	}

	return slog.GroupValue(attrs...)
}

// reflectMapValue - reflect version of MapValue, used in Any wrap, mostly for recursive calls
func reflectMapValue(m reflect.Value) slog.Value {
	if m.IsNil() {
		return slog.StringValue("<nil>")
	}

	if m.Len() == 0 {
		return slog.StringValue("{}")
	}

	attrs := make([]slog.Attr, 0, m.Len())

	var k, v reflect.Value
	for k, v = range m.Seq2() {
		attrs = append(attrs, Any(fmt.Sprintf("%v", k), v))
	}

	return slog.GroupValue(attrs...)
}

// Slice - adds slice values to log as attributes, recursive
func Slice[SL ~[]V, V any](key string, s SL, unpackSingleValue bool) slog.Attr {
	if s == nil {
		return slog.String(key, "<nil>")
	}

	if len(s) == 1 && unpackSingleValue {
		return slog.Any(key, SliceValue(s, unpackSingleValue))
	}

	return slog.Any(key, SliceValue(s, unpackSingleValue).Group())
}

// SliceValue - returns slice group value
func SliceValue[SL ~[]V, V any](s SL, unpackSingleValue bool) slog.Value {
	switch {
	case s == nil:
		return slog.StringValue("<nil>")
	case len(s) == 0:
		if unpackSingleValue {
			return slog.StringValue("")
		}
		return slog.StringValue("[]")
	case len(s) == 1 && unpackSingleValue:
		return slog.AnyValue(s[0])
	}

	attrs := make([]slog.Attr, 0, len(s))
	var (
		v   V
		pos int
	)

	for pos, v = range slices.All(s) {
		attrs = append(attrs, Any(strconv.Itoa(pos), v))
	}

	return slog.GroupValue(attrs...)
}

// reflectSliceValue - reflect version of SliceValue, mostly for recursive calls
func reflectSliceValue(v reflect.Value, unpackSingleValue bool) slog.Value {
	if v.Len() == 0 {
		if unpackSingleValue {
			return slog.StringValue("")
		}
		return slog.StringValue("[]")
	} else if v.Len() == 1 && unpackSingleValue {
		return slog.AnyValue(v.Index(0).Interface())
	}

	attrs := make([]slog.Attr, 0, v.Len())
	var (
		pos, slV reflect.Value
	)

	for pos, slV = range v.Seq2() {
		attrs = append(attrs, Any(strconv.FormatInt(pos.Int(), 10), slV.Interface()))
	}

	return slog.GroupValue(attrs...)
}

// Any - extended version of slog.Any, supports MapValue and SliceValue
func Any(key string, v any) slog.Attr {
	return slog.Any(key, AnyValue(v))
}

// AnyValue - extended version of slog.AnyValue
func AnyValue(v any) slog.Value {
	switch v := v.(type) {
	case string:
		return slog.StringValue(v)
	case int:
		return slog.Int64Value(int64(v))
	case uint:
		return slog.Uint64Value(uint64(v))
	case int64:
		return slog.Int64Value(v)
	case uint64:
		return slog.Uint64Value(v)
	case bool:
		return slog.BoolValue(v)
	case time.Duration:
		return slog.DurationValue(v)
	case time.Time:
		return slog.TimeValue(v)
	case uint8:
		return slog.Uint64Value(uint64(v))
	case uint16:
		return slog.Uint64Value(uint64(v))
	case uint32:
		return slog.Uint64Value(uint64(v))
	case uintptr:
		return slog.Uint64Value(uint64(v))
	case int8:
		return slog.Int64Value(int64(v))
	case int16:
		return slog.Int64Value(int64(v))
	case int32:
		return slog.Int64Value(int64(v))
	case float64:
		return slog.Float64Value(v)
	case float32:
		return slog.Float64Value(float64(v))
	case []slog.Attr:
		return slog.GroupValue(v...)
	case slog.Kind:
		return slog.AnyValue(v)
	case slog.Value:
		return v
	}

	var (
		refV reflect.Value
		ok   bool
	)
	if refV, ok = v.(reflect.Value); !ok {
		refV = reflect.ValueOf(v)
	}

	// unwrap interface for prevent infinite recursion
	if refV.Kind() == reflect.Interface {
		refV = refV.Elem()
	}

	switch refV.Kind() { // nolint: exhaustive
	case reflect.Pointer:
		if refV.IsNil() {
			return slog.StringValue("<nil>")
		}

		return AnyValue(refV.Elem())
	case reflect.Slice:
		return reflectSliceValue(refV, true)
	case reflect.Map:
		return reflectMapValue(refV)
	default:
	}

	if ok {
		return slog.AnyValue(refV.Interface())
	}

	return slog.AnyValue(v)
}
