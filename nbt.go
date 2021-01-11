package nbt

import (
	"compress/gzip"
	"compress/zlib"
	"errors"
	"io"
)

type tagID byte

const (
	TagEnd tagID = iota
	TagByte
	TagShort
	TagInt
	TagLong
	TagFloat
	TagDouble
	TagByteArray
	TagString
	TagList
	TagCompound
	TagIntArray
	TagLongArray
)

type Compound map[string]interface{}

type List []interface{}

// readFunc describes a function which can decode nbt payload
type readFunc func(io.Reader) (interface{}, error)

type writeFunc func(io.Writer) error

func readFuncFactory(id tagID) readFunc {
	switch id {
	case TagEnd:
		return nil
	case TagByte:
		return readByte
	case TagShort:
		return readShort
	case TagInt:
		return readInt
	case TagLong:
		return readLong
	case TagFloat:
		return readFloat
	case TagDouble:
		return readDouble
	case TagByteArray:
		return readByteArray
	case TagString:
		return readString
	case TagIntArray:
		return readIntArray
	case TagLongArray:
		return readLongArray
	case TagCompound:
		return readCompound
	case TagList:
		return readList
	default:
		return nil
	}
}

// Parse parses nbt from io.Reader
func Parse(r io.Reader) (out Compound, err error) {
	v, err := readByte(r)
	id := tagID(v.(byte))
	if id != TagCompound {
		return nil, errors.New("nbt isn't contained in compound")
	}
	// we use compoundWrapRead because every nbt file is implicitly inside one
	v, err = compoundWrapRead(readCompound)(r)
	if v == nil {
		return nil, errors.New("parse: v can't be nil")
	}
	v = v.(compoundField).value
	if v == nil {
		return nil, errors.New("parse: v can't be nil")
	}
	return v.(Compound), err
}

// ParseGzip parses a gzip compressed nbt input stream
func ParseGzip(r io.Reader) (out Compound, err error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return Parse(gr)
}

// ParseGzip parses a zlib compressed nbt input stream
func ParseZlib(r io.Reader) (out Compound, err error) {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	return Parse(zr)
}
