package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"unsafe"
)

var nativeEndian binary.ByteOrder

func init() {
	var one int16 = 1
	b := (*byte)(unsafe.Pointer(&one))
	if *b == 0 {
		nativeEndian = binary.BigEndian
	} else {
		nativeEndian = binary.LittleEndian
	}
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(v interface{}) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	msgLen := uint32(len(buf))
	if err := binary.Write(e.w, nativeEndian, &msgLen); err != nil {
		return err
	}
	if _, err := e.w.Write(buf); err != nil {
		return err
	}
	return nil
}

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(v interface{}) error {
	var msgLen uint32
	if err := binary.Read(d.r, nativeEndian, &msgLen); err != nil {
		return err
	}
	buf := make([]byte, msgLen)
	if _, err := io.ReadFull(d.r, buf); err != nil {
		return err
	}
	return json.Unmarshal(buf, v)
}
