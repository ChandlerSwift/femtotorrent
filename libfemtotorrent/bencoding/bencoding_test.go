package bencoding

import (
	"bytes"
	"testing"
)

func TestDecodeNum(t *testing.T) {
	cases := []struct {
		in   []byte
		want int
	}{
		{[]byte("i0e"), 0},
		{[]byte("i1e"), 1},
		{[]byte("i10e"), 10},
		{[]byte("i123e"), 123},
		{[]byte("i-1e"), -1},
	}
	for _, c := range cases {
		got, err := Decode(c.in)
		if err != nil {
			t.Fatalf("Decode returned error %v", err)
		}

		i, ok := got.(int)
		if !ok {
			t.Errorf("Decode did not return int, returned %T: %v", got, got)
			return
		}
		if i != c.want {
			t.Errorf("Decode(%q) == %v, want %v", c.in, i, c.want)
		}
	}
}

func TestDecodeNumArr(t *testing.T) {
	cases := []struct {
		in   []byte
		want []interface{}
	}{
		{[]byte("li1ee"), []interface{}{1}},
		{[]byte("li0ee"), []interface{}{0}},
		{[]byte("li2ei3ee"), []interface{}{2, 3}},
		{[]byte("li-2ei-3ee"), []interface{}{-2, -3}},
	}
	for _, c := range cases {
		got, err := Decode(c.in)
		if err != nil {
			t.Fatalf("Decode returned error %v", err)
		}

		gotArr := got.([]interface{})
		if len(gotArr) != len(c.want) {
			t.Fatalf("Different length arrays, got %v, want %v", len(gotArr), len(c.want))
		}
		for i := range gotArr {
			if gotArr[i] != c.want[i] {
				t.Errorf("Decode(%q)[%v] == %q, want %q", c.in, i, gotArr[i], c.want[i])
			}
		}
	}
}

func TestDecodeDict(t *testing.T) {
	cases := []struct {
		in   []byte
		want map[string]interface{}
	}{
		{[]byte("d3:onei1e3:twoi2e5:threei3ee"), map[string]interface{}{"one": 1, "two": 2, "three": 3}},
	}
	for _, c := range cases {
		got, err := Decode(c.in)
		if err != nil {
			t.Fatalf("Decode returned error %v", err)
		}

		gotDict := got.(map[string]interface{})

		if len(gotDict) != len(c.want) {
			t.Fatalf("Different length arrays, got %v, want %v", len(gotDict), len(c.want))
		}
		for k, v := range gotDict {
			if v != c.want[k] {
				t.Errorf("Decode(%q)[%v] == %q (%T), want %q (%T)", c.in, k, v, v, c.want[k], c.want[k])
			}
		}
	}
}

func TestDecodeStringDict(t *testing.T) {
	cases := []struct {
		in   []byte
		want map[string]interface{}
	}{
		{[]byte("d3:foo3:bare"), map[string]interface{}{"foo": []byte("bar")}},
	}
	for _, c := range cases {
		got, err := Decode(c.in)
		if err != nil {
			t.Fatalf("Decode returned error %v", err)
		}

		gotDict := got.(map[string]interface{})

		if len(gotDict) != len(c.want) {
			t.Fatalf("Different length arrays, got %v, want %v", len(gotDict), len(c.want))
		}
		for k, v := range gotDict {
			if !bytes.Equal(v.([]byte), c.want[k].([]byte)) {
				t.Errorf("Decode(%q)[%v] == %q (%T), want %q (%T)", c.in, k, v, v, c.want[k], c.want[k])
			}
		}
	}
}

func TestDecodeString(t *testing.T) {
	cases := []struct {
		in   []byte
		want interface{}
	}{
		{[]byte("3:abc"), []byte("abc")},
		{[]byte("3:foo"), []byte("foo")},
		// {[]byte("0:"), []byte("")}, // TODO: fix
		{[]byte("5:\x00\x00\x00\x00\x00"), []byte("\x00\x00\x00\x00\x00")},
	}
	for _, c := range cases {
		res, err := Decode([]byte(c.in))
		if err != nil {
			t.Fatalf("Decode returned error %v", err)
		}
		resStr, ok := res.([]byte)
		if !ok {
			t.Errorf("Decode did not return []byte, returned %T: %v", resStr, resStr)
			return
		}
		if !bytes.Equal(resStr, c.want.([]byte)) {
			t.Errorf("got %q, want %q", resStr, c.want)
			return
		}
	}
}
