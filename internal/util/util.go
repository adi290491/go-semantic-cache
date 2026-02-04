package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

func Float32SliceToBytes(floats []float32) []byte {
	buf := new(bytes.Buffer)

	for _, f := range floats {
		err := binary.Write(buf, binary.LittleEndian, f)
		if err != nil {
			panic(err)
		}
	}

	return buf.Bytes()
}

func CreateQueryHash(key string) string {
	h := sha256.New()
	_, err := h.Write([]byte(key))
	if err != nil {
		panic(err)
	}

	return string(h.Sum(nil))
}
