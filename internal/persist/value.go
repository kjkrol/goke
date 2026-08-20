package persist

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"unsafe"
)

// EncodeValue writes the value of type t located at ptr to w, following the
// same field-by-field rule comp.ValidateEncodable enforces at RegComp: a
// type implementing encoding.BinaryMarshaler is written as a length-prefixed
// opaque blob (checked first, at every level); a string is written as a
// length-prefixed UTF-8 blob; a struct or fixed-size array is written
// field-by-field / element-by-element; a bool or numeric kind is written
// directly. t is assumed already validated — anything else panics.
func EncodeValue(w io.Writer, t reflect.Type, ptr unsafe.Pointer) error {
	return encodeValue(w, reflect.NewAt(t, ptr).Elem())
}

// DecodeValue reads a value of type t from r into the memory at ptr,
// mirroring EncodeValue's rule exactly.
func DecodeValue(r io.Reader, t reflect.Type, ptr unsafe.Pointer) error {
	return decodeValue(r, reflect.NewAt(t, ptr).Elem())
}

func encodeValue(w io.Writer, v reflect.Value) error {
	if m, ok := asBinaryMarshaler(v); ok {
		data, err := m.MarshalBinary()
		if err != nil {
			return err
		}
		if err := writeBytes(w, data); err != nil {
			return err
		}
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		return writeBytes(w, []byte(v.String()))
	case reflect.Bool:
		var b [1]byte
		if v.Bool() {
			b[0] = 1
		}
		_, err := w.Write(b[:])
		return err
	case reflect.Int8:
		return writeUint8(w, uint8(v.Int()))
	case reflect.Int16:
		return writeUint16(w, uint16(v.Int()))
	case reflect.Int32:
		return writeUint32(w, uint32(v.Int()))
	case reflect.Int, reflect.Int64:
		return writeUint64(w, uint64(v.Int()))
	case reflect.Uint8:
		return writeUint8(w, uint8(v.Uint()))
	case reflect.Uint16:
		return writeUint16(w, uint16(v.Uint()))
	case reflect.Uint32:
		return writeUint32(w, uint32(v.Uint()))
	case reflect.Uint, reflect.Uint64:
		return writeUint64(w, v.Uint())
	case reflect.Float32:
		return writeUint32(w, math.Float32bits(float32(v.Float())))
	case reflect.Float64:
		return writeUint64(w, math.Float64bits(v.Float()))
	case reflect.Complex64:
		c := v.Complex()
		if err := writeUint32(w, math.Float32bits(float32(real(c)))); err != nil {
			return err
		}
		return writeUint32(w, math.Float32bits(float32(imag(c))))
	case reflect.Complex128:
		c := v.Complex()
		if err := writeUint64(w, math.Float64bits(real(c))); err != nil {
			return err
		}
		return writeUint64(w, math.Float64bits(imag(c)))
	case reflect.Array:
		for i := range v.Len() {
			if err := encodeValue(w, v.Index(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		for i := range v.NumField() {
			if err := encodeValue(w, v.Field(i)); err != nil {
				return err
			}
		}
		return nil
	default:
		panic(fmt.Sprintf("persist: unencodable kind %s reached the encoder — comp.ValidateEncodable should have rejected this type at RegComp", v.Kind()))
	}
}

func decodeValue(r io.Reader, v reflect.Value) error {
	if u, ok := asBinaryUnmarshaler(v); ok {
		data, err := readBytes(r)
		if err != nil {
			return err
		}
		return u.UnmarshalBinary(data)
	}

	switch v.Kind() {
	case reflect.String:
		data, err := readBytes(r)
		if err != nil {
			return err
		}
		v.SetString(string(data))
		return nil
	case reflect.Bool:
		var b [1]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return err
		}
		v.SetBool(b[0] != 0)
		return nil
	case reflect.Int8:
		n, err := readUint8(r)
		v.SetInt(int64(int8(n)))
		return err
	case reflect.Int16:
		n, err := readUint16(r)
		v.SetInt(int64(int16(n)))
		return err
	case reflect.Int32:
		n, err := readUint32(r)
		v.SetInt(int64(int32(n)))
		return err
	case reflect.Int, reflect.Int64:
		n, err := readUint64(r)
		v.SetInt(int64(n))
		return err
	case reflect.Uint8:
		n, err := readUint8(r)
		v.SetUint(uint64(n))
		return err
	case reflect.Uint16:
		n, err := readUint16(r)
		v.SetUint(uint64(n))
		return err
	case reflect.Uint32:
		n, err := readUint32(r)
		v.SetUint(uint64(n))
		return err
	case reflect.Uint, reflect.Uint64:
		n, err := readUint64(r)
		v.SetUint(n)
		return err
	case reflect.Float32:
		n, err := readUint32(r)
		v.SetFloat(float64(math.Float32frombits(n)))
		return err
	case reflect.Float64:
		n, err := readUint64(r)
		v.SetFloat(math.Float64frombits(n))
		return err
	case reflect.Complex64:
		reBits, err := readUint32(r)
		if err != nil {
			return err
		}
		imBits, err := readUint32(r)
		if err != nil {
			return err
		}
		v.SetComplex(complex(float64(math.Float32frombits(reBits)), float64(math.Float32frombits(imBits))))
		return nil
	case reflect.Complex128:
		reBits, err := readUint64(r)
		if err != nil {
			return err
		}
		imBits, err := readUint64(r)
		if err != nil {
			return err
		}
		v.SetComplex(complex(math.Float64frombits(reBits), math.Float64frombits(imBits)))
		return nil
	case reflect.Array:
		for i := range v.Len() {
			if err := decodeValue(r, v.Index(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		for i := range v.NumField() {
			if err := decodeValue(r, v.Field(i)); err != nil {
				return err
			}
		}
		return nil
	default:
		panic(fmt.Sprintf("persist: unencodable kind %s reached the decoder — comp.ValidateEncodable should have rejected this type at RegComp", v.Kind()))
	}
}

func asBinaryMarshaler(v reflect.Value) (encoding.BinaryMarshaler, bool) {
	if !v.CanAddr() {
		return nil, false
	}
	m, ok := v.Addr().Interface().(encoding.BinaryMarshaler)
	return m, ok
}

func asBinaryUnmarshaler(v reflect.Value) (encoding.BinaryUnmarshaler, bool) {
	if !v.CanAddr() {
		return nil, false
	}
	u, ok := v.Addr().Interface().(encoding.BinaryUnmarshaler)
	return u, ok
}

// writeBytes writes a uint32 length prefix followed by data.
func writeBytes(w io.Writer, data []byte) error {
	if err := writeUint32(w, uint32(len(data))); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}

// readBytes reads a uint32 length prefix followed by that many bytes.
func readBytes(r io.Reader) ([]byte, error) {
	n, err := readUint32(r)
	if err != nil {
		return nil, err
	}
	data := make([]byte, n)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	return data, nil
}

func writeUint8(w io.Writer, v uint8) error {
	_, err := w.Write([]byte{v})
	return err
}

func readUint8(r io.Reader) (uint8, error) {
	var buf [1]byte
	_, err := io.ReadFull(r, buf[:])
	return buf[0], err
}

func writeUint16(w io.Writer, v uint16) error {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

func readUint16(r io.Reader) (uint16, error) {
	var buf [2]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf[:]), nil
}

func writeUint32(w io.Writer, v uint32) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

func readUint32(r io.Reader) (uint32, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf[:]), nil
}

func writeUint64(w io.Writer, v uint64) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

func readUint64(r io.Reader) (uint64, error) {
	var buf [8]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buf[:]), nil
}
