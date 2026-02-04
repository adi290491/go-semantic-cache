package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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

func HashQuery(key string) string {
	hash := sha256.Sum256([]byte(key))

	return hex.EncodeToString(hash[:])
}
