package comp

import (
	"encoding"
	"fmt"
	"reflect"
)

var (
	binaryMarshalerType   = reflect.TypeFor[encoding.BinaryMarshaler]()
	binaryUnmarshalerType = reflect.TypeFor[encoding.BinaryUnmarshaler]()
)

// implementsBinaryCodec reports whether t implements both
// encoding.BinaryMarshaler and encoding.BinaryUnmarshaler, checked via *t so
// value- and pointer-receiver methods both count (a pointer's method set is
// a superset of its base type's).
func implementsBinaryCodec(t reflect.Type) bool {
	pt := reflect.PointerTo(t)
	return pt.Implements(binaryMarshalerType) && pt.Implements(binaryUnmarshalerType)
}

// ValidateEncodable walks t recursively (including t itself) and returns an
// error naming the first field that cannot be encoded. A field is encodable
// if it is a bool, a numeric kind, a string, a struct, a fixed-size array,
// or implements encoding.BinaryMarshaler and encoding.BinaryUnmarshaler
// (checked first, at every level — such a type is treated as opaque and its
// own fields are not inspected). Everything else — Ptr, UnsafePointer,
// Slice, Map, Interface, Chan, Func — is rejected.
func ValidateEncodable(t reflect.Type) error {
	return validateEncodable(t, "")
}

func validateEncodable(t reflect.Type, fieldPath string) error {
	if implementsBinaryCodec(t) {
		return nil
	}
	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return nil
	case reflect.Array:
		return validateEncodable(t.Elem(), fieldPath+"[]")
	case reflect.Struct:
		for i := range t.NumField() {
			f := t.Field(i)
			if err := validateEncodable(f.Type, subPath(fieldPath, f.Name)); err != nil {
				return err
			}
		}
		return nil
	default:
		if fieldPath == "" {
			return fmt.Errorf("comp: type %s (kind %s) is not encodable: implement encoding.BinaryMarshaler and encoding.BinaryUnmarshaler, or avoid pointers, slices, maps, interfaces, channels, and funcs in components", t, t.Kind())
		}
		return fmt.Errorf("comp: field %s has type %s (kind %s) — not encodable: implement encoding.BinaryMarshaler and encoding.BinaryUnmarshaler, or avoid pointers, slices, maps, interfaces, channels, and funcs in components", fieldPath, t, t.Kind())
	}
}

// OffChunkFields walks t recursively (including t itself) and returns the
// dotted paths of every field that requires a dereference outside the
// archetype's contiguous chunk memory during iteration: a string, or a type
// resolved via encoding.BinaryMarshaler/BinaryUnmarshaler. Assumes t is
// already known to be encodable — call ValidateEncodable first.
func OffChunkFields(t reflect.Type) []string {
	var out []string
	collectOffChunkFields(t, "", &out)
	return out
}

func collectOffChunkFields(t reflect.Type, fieldPath string, out *[]string) {
	if implementsBinaryCodec(t) {
		*out = append(*out, displayPath(fieldPath))
		return
	}
	switch t.Kind() {
	case reflect.String:
		*out = append(*out, displayPath(fieldPath))
	case reflect.Array:
		collectOffChunkFields(t.Elem(), fieldPath+"[]", out)
	case reflect.Struct:
		for i := range t.NumField() {
			f := t.Field(i)
			collectOffChunkFields(f.Type, subPath(fieldPath, f.Name), out)
		}
	}
}

func subPath(base, name string) string {
	if base == "" {
		return name
	}
	return base + "." + name
}

func displayPath(fieldPath string) string {
	if fieldPath == "" {
		return "(itself)"
	}
	return fieldPath
}
