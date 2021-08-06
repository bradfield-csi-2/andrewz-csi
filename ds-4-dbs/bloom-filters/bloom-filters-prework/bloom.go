package main

import (
	"encoding/binary"
  //"errors"
  "unsafe"
	//	"fmt"
	//"github.com/spaolacci/murmur3"
	"hash/fnv"
	//"leb.io/hashland/jenkins"
)


const BitPositionMask = 0x03f
const BucketShift = 6

type bloomFilter interface {
	add(item string)

	// `false` means the item is definitely not in the set
	// `true` means the item might be in the set
	maybeContains(item string) bool

	// Number of bytes used in any underlying storage
	memoryUsage() int
}

type aBloomFilter struct {
	data    []uint64
	nHashes uint64
  nBuckets uint64
	nBits   uint64
  bitsPerKey uint64
}

func (b *aBloomFilter) getBitPositions(item string) []uint64 {
  fnvHash := fnv.New64()

	// The basic way to convert `item` into a byte array is as follows:
	//
	//     bytes := []byte(item)
	//
	// However, this `unsafe.Pointer` hack lets us avoid copying data and
	// get a slight speed-up by directly referencing the underlying bytes
	// of `item`.

	fnvHash.Write(*(*[]byte)(unsafe.Pointer(&item)))
	fnvSum := fnvHash.Sum64()
	fnvSum1 := fnvSum & 0xffffffff
	fnvSum2 := fnvSum >> 32
  bitPositions := make([]uint64, b.nHashes)
	for i := uint64(0); i < b.nHashes; i++ {
		bitPositions[i] = (fnvSum1 + fnvSum2*i) % b.nBits
  }
  return bitPositions[:b.nHashes]
}


func newBloomFilter(nItems, bitsPerKey  uint64) *aBloomFilter {
  nBits := nItems * bitsPerKey
  if nBits < 64 {
    nBits = 64
  }

  nBuckets := (nBits + 63) / 64

  nBits = nBuckets * 64

  nHashes := uint64(float64(bitsPerKey) * 0.69)

  if nHashes < 1 {
    nHashes = 1
  }

  if nHashes > 30 {
    nHashes = 30
  }

	return &aBloomFilter{
		data:    make([]uint64, nBuckets),
		nHashes: nHashes,
    nBuckets: nBuckets,
		nBits:    nBits,
    bitsPerKey: bitsPerKey,
	}
}

func (b *aBloomFilter) add(item string) {
	positions := b.getBitPositions(item)

	for _, position := range positions {
	  //bucket := position >> 6
		idx, mask := position >> BucketShift, uint64(1) << (position & BitPositionMask)
		//idx, mask := (hash%uint64(b.bits)) >> 3 , byte(hash&0x07)
		b.data[idx] = b.data[idx] | mask
	}

}

func (b *aBloomFilter) maybeContains(item string) bool {

	positions := b.getBitPositions(item)

	for _, position := range positions {
    //bucket := (hash%uint64(b.bits))
		idx, mask := position >> BucketShift, uint64(1) << (position & BitPositionMask)

		//idx, mask := bucket >> 3, byte(1 << (bucket&0x07))
		if (b.data[idx] & mask) == 0 {
			return false
		}
	}

	return true
}

func (b *aBloomFilter) memoryUsage() int {
	return binary.Size(b.data)
}
