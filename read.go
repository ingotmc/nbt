package nbt

import (
	"encoding/binary"
	"errors"
	"io"
)

func readByte(reader io.Reader) (interface{}, error) {
	b := []byte{0x00}
	_, err := reader.Read(b)
	return b[0], err
}

func binaryRead(r io.Reader, s interface{}) error {
	return binary.Read(r, binary.BigEndian, s)
}

func readShort(r io.Reader) (interface{}, error) {
	s := int16(0x00)
	err := binaryRead(r, &s)
	return s, err
}

func readInt(r io.Reader) (interface{}, error) {
	i := int32(0x00)
	err := binaryRead(r, &i)
	return i, err
}

func readLong(r io.Reader) (interface{}, error) {
	l := int64(0x00)
	err := binaryRead(r, &l)
	return l, err
}

func readFloat(r io.Reader) (interface{}, error) {
	l := float32(0x00)
	err := binaryRead(r, &l)
	return l, err
}

func readDouble(r io.Reader) (interface{}, error) {
	l := float64(0x00)
	err := binaryRead(r, &l)
	return l, err
}

func readByteArray(r io.Reader) (interface{}, error) {
	length, err := readInt(r)
	l := length.(int32)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, int(l))
	_, err = io.ReadFull(r, buf)
	return buf, err
}

func readString(r io.Reader) (interface{}, error) {
	length, err := readShort(r)
	if err != nil {
		return nil, err
	}
	l := uint16(length.(int16))
	buf := make([]byte, int(l))
	_, err = io.ReadFull(r, buf)
	return string(buf), err
}

func readIntArray(r io.Reader) (interface{}, error) {
	length, err := readInt(r)
	if err != nil {
		return nil, err
	}
	l := length.(int32)
	buf := make([]int32, l)
	for i := range buf {
		var v interface{}
		v, err = readInt(r)
		if err != nil {
			return buf, err
		}
		buf[i] = v.(int32)
	}
	return buf, err
}

func readLongArray(r io.Reader) (interface{}, error) {
	length, err := readInt(r)
	if err != nil {
		return nil, err
	}
	l := length.(int32)
	buf := make([]int64, l)
	for i := range buf {
		var v interface{}
		v, err = readLong(r)
		if err != nil {
			return buf, err
		}
		buf[i] = v.(int64)
	}
	return buf, err
}

func readCompound(r io.Reader) (interface{}, error) {
	out := make(Compound)
	var err error
	for {
		v, err := readByte(r)
		if err != nil {
			break
		}
		id := v.(byte)
		rf := readFuncFactory(tagID(id))
		if rf == nil {
			// TagEnd
			break
		}
		v, err = compoundWrapRead(rf)(r)
		if err != nil {
			break
		}
		field := v.(compoundField)
		out[field.name] = field.value
	}
	return out, err
}

type compoundField struct {
	name  string
	value interface{}
}

func compoundWrapRead(rf readFunc) readFunc {
	return func(r io.Reader) (interface{}, error) {
		n, err := readString(r)
		if err != nil {
			return nil, err
		}
		name := n.(string)
		v, err := rf(r)
		if err != nil {
			return nil, err
		}
		return compoundField{
			name:  name,
			value: v,
		}, nil
	}
}

func readList(r io.Reader) (interface{}, error) {
	v, err := readByte(r)
	if err != nil {
		return nil, err
	}
	id := v.(byte)
	rf := readFuncFactory(tagID(id))
	v, err = readInt(r)
	if err != nil {
		return nil, err
	}
	length := v.(int32)
	if length != 0 && rf == nil {
		return nil, errors.New("rf can't be nil if length isn't 0")
	}
	list := make([]interface{}, length)
	for i := range list {
		v, err := rf(r)
		if err != nil {
			break
		}
		list[i] = v
	}
	return list, err
}
