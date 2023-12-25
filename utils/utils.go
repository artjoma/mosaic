package utils

import (
	cp "crypto/rand"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/zeebo/blake3"
	"math/rand"
	"mosaic/types"
)

/*
	Global utils
*/

const TestPortRangeFrom = 6751
const TestPortRangeTo = 6760

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandomString(count int) string {
	b := make([]byte, count)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func RandomBytes(count int) []byte {
	buf := make([]byte, count)
	if _, err := cp.Read(buf); err != nil {
		panic(err)
	}
	return buf
}

// ToBytes MsgPack serialization
func ToBytes(obj interface{}) ([]byte, error) {
	return msgpack.Marshal(obj)
}

// FromBytes MsgPack serialization
func FromBytes(data []byte, obj interface{}) error {
	return msgpack.Unmarshal(data, obj)
}

func U64tob(val uint64) []byte {
	r := make([]byte, 8)
	for i := uint64(0); i < 8; i++ {
		r[i] = byte((val >> (i * 8)) & 0xff)
	}
	return r
}

func Btou64(val []byte) uint64 {
	r := uint64(0)
	for i := uint64(0); i < 8; i++ {
		r |= uint64(val[i]) << (8 * i)
	}
	return r
}

func U32tob(val uint32) []byte {
	r := make([]byte, 4)
	for i := uint32(0); i < 4; i++ {
		r[i] = byte((val >> (8 * i)) & 0xff)
	}
	return r
}

func Btou32(val []byte) uint32 {
	r := uint32(0)
	for i := uint32(0); i < 4; i++ {
		r |= uint32(val[i]) << (8 * i)
	}
	return r
}

func HashFrom(data []byte) types.H256 {
	hasher := blake3.New()
	if _, err := hasher.Write(data); err != nil {
		panic(err)
	}
	return hasher.Sum(nil)
}
