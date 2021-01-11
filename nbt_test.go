package nbt

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)


const bigtestPath = "testdata/bigtest.nbt"

func TestCodec(t *testing.T) {
	f, err := os.Open(bigtestPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	origCmpd, err := ParseGzip(f)
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer(make([]byte, 0, 400))
	err = EncodeGzip(origCmpd, buf)
	if err != nil {
		t.Fatal(err)
	}
	cmpd, err := ParseGzip(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(origCmpd, cmpd) {
		t.Error("compound wasn't preserved by decoding and re-encoding (not necessarily a bug, might be DeepEqual fault's)")
	}
}