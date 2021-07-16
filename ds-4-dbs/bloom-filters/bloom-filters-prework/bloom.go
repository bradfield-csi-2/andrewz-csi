package main

import (
	"encoding/binary"
	//	"fmt"
	//"github.com/spaolacci/murmur3"
	"hash/fnv"
  //"leb.io/hashland/jenkins"
)

var S int = 2
var Seeds [9]uint32 = [9]uint32{101,65537,223,263,277,307,433,509,577}

type bloomFilter interface {
	add(item string)

	// `false` means the item is definitely not in the set
	// `true` means the item might be in the set
	maybeContains(item string) bool

	// Number of bytes used in any underlying storage
	memoryUsage() int
}

type trivialBloomFilter struct {
	data []uint64
}

func newTrivialBloomFilter() *trivialBloomFilter {
	return &trivialBloomFilter{
		data: make([]uint64, 1024),
	}
}

func (b *trivialBloomFilter) add(item string) {
	// Do nothing
}

func (b *trivialBloomFilter) maybeContains(item string) bool {
	// Technically, any item "might" be in the set
	return true
}

func (b *trivialBloomFilter) memoryUsage() int {
	return binary.Size(b.data)
}

//AZ Bloomfilter
type azBloomFilter struct {
	data []byte
}

type azBloomIdxMaskPair struct {
	index int
	mask  byte
}

func getHashIdxMaskPairs(item string) [6]azBloomIdxMaskPair {
	pairs := [6]azBloomIdxMaskPair{}//make([]azBloomIdxMaskPair, 0)
  /*
	result := murmur3.Sum32([]byte(item))
	idx := int((result & 0x0fffff) >> 3)
	off := result & 0x07
	offmask := byte(1 << int(off))
  pairs[0] = azBloomIdxMaskPair{idx, offmask}
	//pairs = append(pairs, azBloomIdxMaskPair{idx, offmask})

	fnv1 := fnv.New32()
	fnv1.Write([]byte(item))
	result = fnv1.Sum32()
	idx = int((result & 0x0fffff) >> 3)
	off = result & 0x07
	offmask = byte(1 << int(off))
	//pairs = append(pairs, azBloomIdxMaskPair{idx, offmask})
  pairs[1] = azBloomIdxMaskPair{idx, offmask}

	//jenk3 := jenkins.New(100)
	//jenk3.Write([]byte(item))
  result = jenkins.Sum32([]byte(item), 0)
	idx = int((result & 0x0fffff) >> 3)
	off = result & 0x07
	offmask = byte(1 << int(off))
	//pairs = append(pairs, azBloomIdxMaskPair{idx, offmask})
  pairs[2] = azBloomIdxMaskPair{idx, offmask}
*/
  k := 1
  for i := 0; i < k; i++ {

    //result := jenkins.Sum32([]byte(item), Seeds[i])
  	fnv1 := fnv.New64()
	  fnv1.Write([]byte(item))
    result := fnv1.Sum64()
	
    hold1 := result & 0x0fffff
    idx := int((result & 0x0fffff) >> 3)
    off := result & 0x07
    offmask := byte(1 << int(off))
	  //pairs = append(pairs, azBloomIdxMaskPair{idx, offmask})
    pairs[k*i + 0] = azBloomIdxMaskPair{idx, offmask}

    //result = result >> 16//jenkins.Sum32([]byte(item), Seeds[i])
    hold2 := result >> 20
    idx = int((hold2 & 0x0fffff) >> 3)
    off = hold2 & 0x07
    offmask = byte(1 << int(off))
	  //pairs = append(pairs, azBloomIdxMaskPair{idx, offmask})
    pairs[k*i + 1] = azBloomIdxMaskPair{idx, offmask}

    hold3 := result >> 40
    idx = int((hold3 & 0x0fffff) >> 3)
    off = hold3 & 0x07
    offmask = byte(1 << int(off))
	  //pairs = append(pairs, azBloomIdxMaskPair{idx, offmask})
    pairs[k*i + 2] = azBloomIdxMaskPair{idx, offmask}


    hold4 := hold1 ^ hold2
    idx = int((hold4 & 0x0fffff) >> 3)
    off = hold4 & 0x07
    offmask = byte(1 << int(off))
	  //pairs = append(pairs, azBloomIdxMaskPair{idx, offmask})
    pairs[k*i + 3] = azBloomIdxMaskPair{idx, offmask}

    hold5 := hold1 ^ hold3
    idx = int((hold5 & 0x0fffff) >> 3)
    off = hold5 & 0x07
    offmask = byte(1 << int(off))
	  //pairs = append(pairs, azBloomIdxMaskPair{idx, offmask})
    pairs[k*i + 4] = azBloomIdxMaskPair{idx, offmask}

    hold6 := hold2 ^ hold3
    idx = int((hold6 & 0x0fffff) >> 3)
    off = hold6 & 0x07
    offmask = byte(1 << int(off))
	  //pairs = append(pairs, azBloomIdxMaskPair{idx, offmask})
    pairs[k*i + 5] = azBloomIdxMaskPair{idx, offmask}



 
  }

	return pairs
}

func newAZBloomFilter() *azBloomFilter {
	return &azBloomFilter{
		data: make([]byte, 131072),
	}
}

func (b *azBloomFilter) add(item string) {
	/*
		result := murmur3.Sum32([]byte(item))
		idx := (result & 0x0fffff) >> 3
		off := result & 0x07
	  offmask := byte(1 << int(off))
		b.data[idx] = b.data[idx] | offmask
	*/
	pairs := getHashIdxMaskPairs(item)

	for i, hpair := range pairs {
    if i > 9 {
      break
    }
		idx, mask := hpair.index, hpair.mask
		b.data[idx] = b.data[idx] | mask
	}

}

func (b *azBloomFilter) maybeContains(item string) bool {
	/*
		result := murmur3.Sum32([]byte(item))
		idx := (result & 0x0fffff) >> 3
		off := result & 0x07
	  offmask := byte(1 << int(off))
		return (b.data[idx] & offmask) != 0
	  pairs := getHashIdxMaskPairs(item)
	*/
	pairs := getHashIdxMaskPairs(item)

	for i, hpair := range pairs {
    if i > 9 {
      break
    }
		idx, mask := hpair.index, hpair.mask
		if (b.data[idx] & mask) == 0 {
			return false
		}
	}

	return true
}

func (b *azBloomFilter) memoryUsage() int {
	return binary.Size(b.data)
}
