package main

import (
	"encoding/binary"
  "errors"
	//	"fmt"
	//"github.com/spaolacci/murmur3"
	"hash/fnv"
	//"leb.io/hashland/jenkins"
)

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
  morek   bool
}

func (b *aBloomFilter) getHashes(item string) []uint64 {
  k := b.nHashes
  if b.morek {
    k *=3
  }
	hashes := make([]uint64, k)

	fnv1 := fnv.New64()
	for i := 0; i < b.nHashes; i++ {
		fnv1.Write([]byte{byte(i)})
		fnv1.Write([]byte(item))
		h := fnv1.Sum64()
		//hashes[b.bHashes*0+i] = h

    if b.morek {
      h1 := h >> 32
      h = h & 0x0ffffffff
      h2 := h1 ^ h
      hashes[3+i] = h1
      hashes[6+i] = h2
    }
    hashes[i] = h
	}

	return hashes
}

func (b *aBloomFilter) getMoreHashes(hash64 uint64) ( uint32, uint32) {
  return 0,0
}



func newBloomFilter(nBytes, nHashes int) *aBloomFilter {
	return &aBloomFilter{
		data:    make([]byte, nBytes),
		nHashes: nHashes,
    buckets: nBytes,
		bits:    nBytes * 8,
    morek: false,
	}
}

func newBloomFilterMoreK(nBytes, nHashes int)(*aBloomFilter, error) {
  if (nBytes << 3) >= 1 << 32 {
    return nil, errors.New("size too big for this feature")
  }
	return &aBloomFilter{
		data:    make([]byte, nBytes),
		nHashes: nHashes,
    buckets: nBytes,
		bits:    nBytes * 8,
    morek: true,
	}, nil
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
