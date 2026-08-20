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

// promotedCodecHazard reports whether t's BinaryMarshaler/BinaryUnmarshaler
// implementation is Go method promotion from an embedded (anonymous) field,
// while t also has other fields alongside it — the embedded type's
// MarshalBinary has no way to know about those sibling fields, so they'd be
// silently dropped every Save. Returns the embedded field's name for the
// error message. A struct whose ONLY field is the embedded type is safe
// (nothing else to lose); anything named directly on t (not promoted) is
// also safe, since Go resolves a type's own method over a promoted one.
func promotedCodecHazard(t reflect.Type) (embeddedField string, hazard bool) {
	if t.Kind() != reflect.Struct || t.NumField() <= 1 {
		return "", false
	}
	for i := range t.NumField() {
		f := t.Field(i)
		if f.Anonymous && implementsBinaryCodec(f.Type) {
			return f.Name, true
		}
	}
	return "", false
}

// ValidateEncodable walks t recursively (including t itself) and returns an
// error naming the first field that cannot be encoded. A field is encodable
// if it is a bool, a numeric kind (including complex64/complex128), a
// string, an exported struct field, a fixed-size array, or implements
// encoding.BinaryMarshaler and encoding.BinaryUnmarshaler (checked first, at
// every level — such a type is treated as opaque and its own fields are not
// inspected). Everything else — Ptr, UnsafePointer, Uintptr, Slice, Map,
// Interface, Chan, Func, and unexported struct fields — is rejected; Uintptr
// is grouped with the pointer kinds because it conventionally holds a raw
// address, not a portable value. Unexported fields are rejected because the
// encoder needs reflect.Value.Interface on every field it visits (to probe
// for BinaryMarshaler), which Go's reflect package refuses for a value
// sourced from an unexported field, regardless of that field's own type.
// A type that only implements BinaryMarshaler/BinaryUnmarshaler via Go
// method promotion from an embedded field, while also declaring other
// fields of its own, is rejected too — the promoted method has no way to
// know about those sibling fields, so they'd be silently dropped every
// Save. See promotedCodecHazard.
func ValidateEncodable(t reflect.Type) error {
	return validateEncodable(t, "")
}

func validateEncodable(t reflect.Type, fieldPath string) error {
	if implementsBinaryCodec(t) {
		if embedded, hazard := promotedCodecHazard(t); hazard {
			what := t.String()
			if fieldPath != "" {
				what = fmt.Sprintf("field %s (type %s)", fieldPath, t)
			}
			return fmt.Errorf("comp: %s implements BinaryMarshaler/BinaryUnmarshaler only via its embedded field %s — its other fields would be silently dropped every Save, since the promoted MarshalBinary only knows about %s's own fields; give %s its own MarshalBinary/UnmarshalBinary covering everything, or stop embedding %s and use a named field instead",
				what, embedded, embedded, t, embedded)
		}
		return nil
	}
	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return nil
	case reflect.Array:
		return validateEncodable(t.Elem(), fieldPath+"[]")
	case reflect.Struct:
		for i := range t.NumField() {
			f := t.Field(i)
			path := subPath(fieldPath, f.Name)
			if !f.IsExported() {
				return fmt.Errorf("comp: field %s is unexported — persist cannot read unexported struct fields (Go's reflect forbids it); export the field, or implement encoding.BinaryMarshaler and encoding.BinaryUnmarshaler on the containing type to treat it as opaque", path)
			}
			if err := validateEncodable(f.Type, path); err != nil {
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
