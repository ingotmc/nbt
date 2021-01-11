package nbt

import (
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"io"
)

func Encode(c Compound, w io.Writer) error {
	_, err := w.Write([]byte{byte(TagCompound)})
	if err != nil {
		return err
	}
	return compoundWrapWrite("", writeCompound(c))(w)
}

func EncodeGzip(c Compound, w io.Writer) error {
	gW := gzip.NewWriter(w)
	defer gW.Close()
	return Encode(c, gW)
}

func EncodeZlib(c Compound, w io.Writer) error {
	zW := zlib.NewWriter(w)
	defer zW.Close()
	return Encode(c, zW)
}

// encodeFactory takes in a value and returns the respective writeFunc and tagID
func encodeFactory(v interface{}) (writeFunc, tagID) {
	switch x := v.(type) {
	case Compound:
		return writeCompound(x), TagCompound
	case List:
		return writeList(x), TagList
	case []interface{}:
		return writeList(x), TagList
	case uint8:
		return writeByte(x), TagByte
	case int8:
		return writeByte(uint8(x)), TagByte
	case uint16:
		return writeShort(x), TagShort
	case int16:
		return writeShort(uint16(x)), TagShort
	case uint32:
		return writeInt(x), TagInt
	case int32:
		return writeInt(uint32(x)), TagInt
	case uint64:
		return writeLong(x), TagLong
	case int64:
		return writeLong(uint64(x)), TagLong
	case int:
		return writeLong(uint64(x)), TagLong
	case uint:
		return writeLong(uint64(x)), TagLong
	case float32:
		return writeFloat(x), TagFloat
	case float64:
		return writeDouble(x), TagDouble
	case []byte:
		return writeByteArray(x), TagByteArray
	case []uint32:
		return writeIntArray(x), TagIntArray
	case []int32:
		data := make([]uint32, len(x))
		for i, v := range x {
			data[i] = uint32(v)
		}
		return writeIntArray(data), TagIntArray
	case []uint64:
		return writeLongArray(x), TagLongArray
	case []int64:
		data := make([]uint64, len(x))
		for i, v := range x {
			data[i] = uint64(v)
		}
		return writeLongArray(data), TagLongArray
	case []uint:
		data := make([]uint64, len(x))
		for i, v := range x {
			data[i] = uint64(v)
		}
		return writeLongArray(data), TagLongArray
	case []int:
		data := make([]uint64, len(x))
		for i, v := range x {
			data[i] = uint64(v)
		}
		return writeLongArray(data), TagLongArray
	case string:
		return writeString(x), TagString
	default:
		return nil, TagEnd
	}
}

func writeByte(b byte) writeFunc {
	return func(w io.Writer) error {
		_, err := w.Write([]byte{b})
		return err
	}
}

func binaryWrite(w io.Writer, s interface{}) error {
	return binary.Write(w, binary.BigEndian, s)
}

func writeShort(s uint16) writeFunc {
	return func(w io.Writer) error {
		return binaryWrite(w, s)
	}
}

func writeInt(i uint32) writeFunc {
	return func(w io.Writer) error {
		return binaryWrite(w, i)
	}
}

func writeLong(l uint64) writeFunc {
	return func(w io.Writer) error {
		return binaryWrite(w, l)
	}
}

func writeFloat(f float32) writeFunc {
	return func(w io.Writer) error {
		return binaryWrite(w, f)
	}
}

func writeDouble(d float64) writeFunc {
	return func(w io.Writer) error {
		return binaryWrite(w, d)
	}
}

func writeByteArray(data []byte) writeFunc {
	return func(w io.Writer) error {
		l := uint32(len(data))
		err := writeInt(l)(w)
		if err != nil {
			return err
		}
		for _, b := range data {
			err = writeByte(b)(w)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func writeString(s string) writeFunc {
	return func(w io.Writer) error {
		l := uint16(len(s))
		err := writeShort(l)(w)
		if err != nil {
			return err
		}
		for _, b := range []byte(s) {
			err = writeByte(b)(w)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func writeIntArray(data []uint32) writeFunc {
	return func(w io.Writer) error {
		l := uint32(len(data))
		err := writeInt(l)(w)
		if err != nil {
			return err
		}
		for _, b := range data {
			err = writeInt(b)(w)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func writeLongArray(data []uint64) writeFunc {
	return func(w io.Writer) error {
		l := uint32(len(data))
		err := writeInt(l)(w)
		if err != nil {
			return err
		}
		for _, b := range data {
			err = writeLong(b)(w)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func writeCompound(c Compound) writeFunc {
	return func(w io.Writer) error {
		for name, value := range c {
			wf, id := encodeFactory(value)
			if wf == nil {
				// TODO: error here somehow
				break
			}
			err := writeByte(byte(id))(w)
			if err != nil {
				return err
			}
			wf = compoundWrapWrite(name, wf)
			err = wf(w)
			if err != nil {
				break
			}
		}
		return writeByte(byte(TagEnd))(w)
	}
}

func compoundWrapWrite(name string, wf writeFunc) writeFunc {
	return func(w io.Writer) error {
		err := writeString(name)(w)
		if err != nil {
			return err
		}
		return wf(w)
	}
}

func writeList(l []interface{}) writeFunc {
	return func(w io.Writer) error {
		length := uint32(len(l))
		if length < 1 {
			return nil
		}
		_, id := encodeFactory(l[0])
		err := writeByte(byte(id))(w)
		if err != nil {
			return err
		}
		err = writeInt(length)(w)
		if err != nil {
			return err
		}
		for _, v := range l {
			wf, _ := encodeFactory(v)
			if wf == nil {
				break
			}
			err = wf(w)
			if err != nil {
				break
			}
		}
		return err
	}
}
