package main

import (
	"bytes"
	"reflect"
	"testing"
)

var streamTest = []interface{}{
	0.1,
	6.28,
	"hello",
	nil,
	true,
	false,
	[]interface{}{"a", "b", "c"},
	map[string]interface{}{"A": "Apple", "C": "Cat"},
	3.14,
}

var streamEncoded = [][]byte{
	[]byte{},
	[]byte{0x03, 0x00, 0x00, 0x00, '0', '.', '1'},
	[]byte{0x04, 0x00, 0x00, 0x00, '6', '.', '2', '8'},
	[]byte{0x07, 0x00, 0x00, 0x00, '"', 'h', 'e', 'l', 'l', 'o', '"'},
	[]byte{0x04, 0x00, 0x00, 0x00, 'n', 'u', 'l', 'l'},
	[]byte{0x04, 0x00, 0x00, 0x00, 't', 'r', 'u', 'e'},
	[]byte{0x05, 0x00, 0x00, 0x00, 'f', 'a', 'l', 's', 'e'},
	[]byte{0x0D, 0x00, 0x00, 0x00, '[', '"', 'a', '"', ',', '"', 'b', '"', ',', '"', 'c', '"', ']'},
	[]byte{0x17, 0x00, 0x00, 0x00, '{', '"', 'A', '"', ':', '"', 'A', 'p', 'p', 'l', 'e', '"', ',', '"', 'C', '"', ':', '"', 'C', 'a', 't', '"', '}'},
	[]byte{0x04, 0x00, 0x00, 0x00, '3', '.', '1', '4'},
}

func TestEncoder(t *testing.T) {
	for i := 0; i <= len(streamTest); i++ {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		for j, v := range streamTest[0:i] {
			if err := enc.Encode(v); err != nil {
				t.Fatalf("encode #%d: %v", j, err)
			}
		}
		if have, want := buf.Bytes(), bytes.Join(streamEncoded[0:i+1], nil); !bytes.Equal(have, want) {
			t.Errorf("encoding %d items: mismatch", i)
			diff(t, have, want)
			break
		}
	}
}

func TestDecoder(t *testing.T) {
	for i := 0; i <= len(streamTest); i++ {
		buf := bytes.NewBuffer(bytes.Join(streamEncoded[0:i+1], nil))
		out := make([]interface{}, i)
		dec := NewDecoder(buf)
		for j := range out {
			if err := dec.Decode(&out[j]); err != nil {
				t.Fatalf("decode #%d/%d: %v", j, i, err)
			}
		}
		if !reflect.DeepEqual(out, streamTest[0:i]) {
			t.Errorf("decoding %d items: mismatch", i)
			for j := range out {
				if !reflect.DeepEqual(out[j], streamTest[j]) {
					t.Errorf("#%d: have %v want %v", j, out[j], streamTest[j])
				}
			}
			break
		}
	}
}

func diff(t *testing.T, a, b []byte) {
	for i := 0; ; i++ {
		if i >= len(a) || i >= len(b) || a[i] != b[i] {
			j := i - 10
			if j < 0 {
				j = 0
			}
			t.Errorf("diverge at %d: «%s» vs «%s»", i, trim(a[j:]), trim(b[j:]))
			return
		}
	}
}

func trim(b []byte) []byte {
	if len(b) > 20 {
		return b[0:20]
	}
	return b
}
