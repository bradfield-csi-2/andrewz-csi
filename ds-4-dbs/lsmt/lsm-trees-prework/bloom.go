package table//main

import (
	"encoding/binary"
  //"errors"
  //
	//	"fmt"
	//"github.com/spaolacci/murmur3"
  "unsafe"
	"hash/fnv"
	//"leb.io/hashland/jenkins"
)

const BitPositionMask = 0x03f
const BucketNumMask = 0xffffffffffffffc0

var numberOfHashFunctions uint64 = 9
var memoryUsageInBits uint64 = 8

type bloomFilter interface {
	add(item string)

	// `false` means the item is definitely not in the set
	// `true` means the item might be in the set
	maybeContains(item string) bool

	// Number of bytes used in any underlying storage
	memoryUsage() int
}

type aBloomFilter struct {
	data    []byte
	nHashes int
  buckets int
	bits    int
  bitsPerKey int
}


func getBitPositions(item string) []uint64 {
	fnvHash := fnv.New64()

	// The hack below was copied from the solution:
	// next time I should be smart enough to come up with it myself:)
	// since I saw in pprof that stringtobyteslice takes 27% of total cpu time

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
  bitPositions := make([]uint64,20)
	for i := uint64(0); i < numberOfHashFunctions; i++ {
		bitPositions[i] = (fnvSum1 + fnvSum2*i) % memoryUsageInBits
	}
	return bitPositions
}

func (b *aBloomFilter) getBitPositions(item string) []uint64 {
	fnvHash := fnv.New64()

  bitPositions := make([]uint64,20)
	fnvHash.Write(*(*[]byte)(unsafe.Pointer(&item)))
	fnvSum := fnvHash.Sum64()
	fnvSum1 := fnvSum & 0xffffffff
	fnvSum2 := fnvSum >> 32
	for i := uint64(0); i < b.nHashes; i++ {
		bitPositions[i] = (fnvSum1 + fnvSum2*i) % b.bits//memoryUsageInBits
	}
	return bitPositions
}



func newBloomFilter(nItems, bitsPerKey int) *aBloomFilter {

	return &aBloomFilter{
		data:    make([]uint64, nBytes),
		nHashes: nHashes,
    buckets: nBytes,
		bits:    nBytes * 64,
    bitsPerKey: bitsPerKey
	}
}


func (b *aBloomFilter) add(item string) {
	hashes := b.getHashes(item)

	for _, hash := range hashes {
	  bucket := (hash%uint64(b.bits))
		idx, mask := bucket >> 3, byte(1 << (bucket&0x07))
		//idx, mask := (hash%uint64(b.bits)) >> 3 , byte(hash&0x07)
		b.data[idx] = b.data[idx] | mask
	}

}

func (b *aBloomFilter) maybeContains(item string) bool {

	hashes := b.getHashes(item)

	for _, hash := range hashes {
    bucket := (hash%uint64(b.bits))
		idx, mask := bucket >> 3, byte(1 << (bucket&0x07))
		if (b.data[idx] & mask) == 0 {
			return false
		}
	}

	return true
}

func (b *aBloomFilter) memoryUsage() int {
	return binary.Size(b.data)
}
