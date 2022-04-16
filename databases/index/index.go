package main

import (
	"encoding/binary"
	"fmt"
)

const (
	VARCHAR = iota
	Int_8
	Int_16
	Int_32
	Int_64
	Uint_8
	Uint_16
	Uint_32
	Uint_64
)

type indexEntry struct {
	key   []byte
	block int64
	slot  int
}

func CompareValues(a, b []byte, f func(a, b []byte) bool) bool {
	return f(a, b)
}

func CompareTyp(a, b []byte, typ int) bool {
	switch typ {
	case VARCHAR:
		return CompareString(a, b)
	case Int_8:
		return CompareInt8(a, b)
	case Int_16:
		return CompareInt16(a, b)
	case Int_32:
		return CompareInt32(a, b)
	case Int_64:
		return CompareInt64(a, b)
	case Uint_8:
		return CompareUint8(a, b)
	case Uint_16:
		return CompareUint16(a, b)
	case Uint_32:
		return CompareUint32(a, b)
	case Uint_64:
		return CompareUint64(a, b)
	default:
		panic("type not handled for compare")
	}
}

func CompareInt64(a, b []byte) bool {
	if len(a) != 8 || len(b) != 8 {
		panic("wrong byte len")
	}
	au := binary.LittleEndian.Uint64(a)
	bu := binary.LittleEndian.Uint64(b)
	return int64(au) < int64(bu)
}

func CompareUint64(a, b []byte) bool {
	if len(a) != 8 || len(b) != 8 {
		panic("wrong byte len")
	}
	au := binary.LittleEndian.Uint64(a)
	bu := binary.LittleEndian.Uint64(b)
	return au < bu
}

func CompareString(a, b []byte) bool {
	return string(a) < string(b)
}

func CompareInt32(a, b []byte) bool {
	if len(a) != 4 || len(b) != 4 {
		panic("wrong byte len")
	}
	au := binary.LittleEndian.Uint32(a)
	bu := binary.LittleEndian.Uint32(b)
	return int32(au) < int32(bu)
}

func CompareUint32(a, b []byte) bool {
	if len(a) != 4 || len(b) != 4 {
		panic("wrong byte len")
	}
	au := binary.LittleEndian.Uint32(a)
	bu := binary.LittleEndian.Uint32(b)
	return au < bu
}

func CompareInt16(a, b []byte) bool {
	if len(a) != 2 || len(b) != 2 {
		panic("wrong byte len")
	}
	au := binary.LittleEndian.Uint16(a)
	bu := binary.LittleEndian.Uint16(b)
	return int16(au) < int16(bu)
}

func CompareUint16(a, b []byte) bool {
	if len(a) != 2 || len(b) != 2 {
		panic("wrong byte len")
	}
	au := binary.LittleEndian.Uint16(a)
	bu := binary.LittleEndian.Uint16(b)
	return au < bu
}

func CompareInt8(a, b []byte) bool {
	if len(a) != 1 || len(b) != 1 {
		panic("wrong byte len")
	}
	return int8(a[0]) < int8(b[0])
}

func CompareUint8(a, b []byte) bool {
	if len(a) != 1 || len(b) != 1 {
		panic("wrong byte len")
	}
	return a[0] < b[0]
}

func test(x []byte) {
	u := binary.LittleEndian.Uint64(x)
	i := int(u)
	fmt.Println(i)
}
