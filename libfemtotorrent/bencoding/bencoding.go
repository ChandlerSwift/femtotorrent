// bencoding implements the encoding used for .torrent files
//
// Strings are length-prefixed base ten followed by a colon and the string. For
// example 4:spam corresponds to 'spam'.
//
// Integers are represented by an 'i' followed by the number in base 10 followed
// by an 'e'. For example i3e corresponds to 3 and i-3e corresponds to -3.
// Integers have no size limitation. i-0e is invalid. All encodings with a
// leading zero, such as i03e, are invalid, other than i0e, which of course
// corresponds to 0.
//
// Lists are encoded as an 'l' followed by their elements (also bencoded)
// followed by an 'e'. For example l4:spam4:eggse corresponds to ['spam',
// 'eggs'].
//
// Dictionaries are encoded as a 'd' followed by a list of alternating keys and
// their corresponding values followed by an 'e'. For example,
// d3:cow3:moo4:spam4:eggse corresponds to {'cow': 'moo', 'spam': 'eggs'} and
// d4:spaml1:a1:bee corresponds to {'spam': ['a', 'b']}. Keys must be strings
// and appear in sorted order (sorted as raw strings, not alphanumerics).
package bencoding

import (
	"bytes"
	"fmt"
	"strconv"
)

func decodeFirstToken(b []byte) ([]byte, interface{}, error) {
	var err error
	switch b[0] {
	case 'l': // list
		b = b[1:] // slurp up the 'l'
		var ret []interface{}
		for b[0] != 'e' {
			var new interface{}
			b, new, err = decodeFirstToken(b)
			if err != nil {
				return nil, nil, err
			}
			ret = append(ret, new)
		}
		if b[0] != 'e' {
			return nil, nil, fmt.Errorf("expected e at end of list, got %v", b[1])
		}
		return b[1:], ret, nil
	case 'i': // int
		b = b[1:] // slurp up the 'i'
		len := 0
		if b[0] == '-' {
			len += 1
		}
		// "Be lenient in what you accept" -- this does allow for leading zeroes
		// which the spec says should not be permissible. TODO: strict mode?
		if bytes.IndexByte([]byte("1234567890"), b[len]) < 0 { // We need at least one byte
			return nil, nil, fmt.Errorf("invalid digit %v", b[len])
		}
		for bytes.IndexByte([]byte("1234567890"), b[len]) >= 0 {
			len += 1
		}
		if b[len] != 'e' {
			return nil, nil, fmt.Errorf("expected e at end of int, got %v", b[len])
		}
		i, err := strconv.Atoi(string(b[:len]))
		return b[len+1:], i, err
	case 'd': // dict
		b = b[1:] // slurp up the 'd'
		ret := make(map[string]interface{})
		for b[0] != 'e' {
			var key, val interface{}
			b, key, err = decodeFirstToken(b)
			if err != nil {
				return nil, nil, err
			}
			b, val, err = decodeFirstToken(b)
			if err != nil {
				return nil, nil, err
			}
			// We can't use []byte as a key because slices are unhashable, and
			// map keys must be hashable. We could mayyyybe conditionally
			// convert only byte slices to strings, or something, but that
			// really muddies up the code. The spec isn't very clear on whether
			// or not dict keys have to be strings, but in practice it seems
			// that they always are, so this (usually) works fine. This _would_
			// cause problems if a key wasn't a valid UTF-8 string, but again,
			// that doesn't seem to be common (as the keys are usually human-
			// readable anyway).
			keyArr, ok := key.([]byte)
			if !ok {
				return nil, nil, fmt.Errorf("expected string key, got %T (%v)", key, key)
			}
			ret[string(keyArr)] = val
		}
		if b[0] != 'e' {
			return nil, nil, fmt.Errorf("expected e at end of list, got %v", b[1])
		}
		return b[1:], ret, nil
	case '1', '2', '3', '4', '5', '6', '7', '8', '9': // string (starting with length)
		lenLen := 0
		for bytes.IndexByte([]byte("1234567890"), b[lenLen]) >= 0 {
			lenLen += 1
		}
		len, err := strconv.Atoi(string(b[:lenLen]))
		if err != nil {
			return nil, nil, err
		}
		b = b[lenLen:]
		if b[0] != ':' {
			return nil, nil, fmt.Errorf("expecting ':', got %v", b)
		}
		b = b[1:]
		return b[len:], b[:len], nil
	}
	return nil, nil, nil
}

func Decode(b []byte) (interface{}, error) {
	remainder, result, err := decodeFirstToken(b)
	if err != nil {
		return nil, err
	} else if len(remainder) != 0 {
		return nil, fmt.Errorf("leftover garbage: %q", remainder)
	}
	return result, nil
}

func Encode(i interface{}) []byte {
	panic("unimplemented")
}
