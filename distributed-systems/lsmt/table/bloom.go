package table

import (
	"encoding/binary"
	"hash/fnv"
	"io"
	"unsafe"
)

const BitPositionMask = 0x03f
const BucketShift = 6

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
	data       []uint64
	nHashes    uint64
	nBuckets   uint64
	nBits      uint64
	bitsPerKey uint64
}

type bloomFilterBlock struct {
	NHashes    uint64
	NBuckets   uint64
	NBits      uint64
	BitsPerKey uint64
}

func (b *aBloomFilter) getBlock() bloomFilterBlock {
	return bloomFilterBlock{
		b.nHashes,
		b.nBuckets,
		b.nBits,
		b.bitsPerKey,
	}
}

func (b *bloomFilterBlock) Write(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, b)
	if err != nil {
		return err
	}
	return nil
}

func (b *bloomFilterBlock) createFilter() aBloomFilter {
	return aBloomFilter{
		nHashes:    b.NHashes,
		nBuckets:   b.NBuckets,
		nBits:      b.NBits,
		bitsPerKey: b.BitsPerKey,
	}
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
	return bitPositions
}

//has some significants iun determining the size and accuracy
//but maybe's don't want ideal accuracy for given size??
//maybe only for a certain size?
func newBloomFilter(nItems, bitsPerKey uint64) *aBloomFilter {
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
		data:       make([]uint64, nBuckets),
		nHashes:    nHashes,
		nBuckets:   nBuckets,
		nBits:      nBits,
		bitsPerKey: bitsPerKey,
	}
}

func (b *aBloomFilter) add(item string) {
	positions := b.getBitPositions(item)

	for _, position := range positions {
		idx, mask := position>>BucketShift, uint64(1)<<(position&BitPositionMask)
		b.data[idx] = b.data[idx] | mask
	}

}

func (b *aBloomFilter) maybeContains(item string) bool {

	positions := b.getBitPositions(item)

	for _, position := range positions {
		idx, mask := position>>BucketShift, uint64(1)<<(position&BitPositionMask)
		if (b.data[idx] & mask) == 0 {
			return false
		}
	}

	return true
}

func (b *aBloomFilter) memoryUsage() int {
	return binary.Size(b.data)
}

func (b *aBloomFilter) Write(w io.Writer) error {
	//compressWriter, err := zlib.NewWriterLevel(w, zlib.BestCompression)
	//defer compressWriter.Close()
	var err error
	compressWriter := w

	if err != nil {
		return err
	}

	block := b.getBlock()

	err = binary.Write(compressWriter, binary.LittleEndian, block)
	if err != nil {
		return err
	}

	err = binary.Write(compressWriter, binary.LittleEndian, b.data)
	if err != nil {
		return err
	}
	/*
		err = binary.Write(compressWriter, binary.LittleEndian, b.bytes)
		if err != nil {
			return err
		}

		//err = compressWriter.Flush()
		if err != nil {
			return err
		}
	*/
	return nil
}

func loadBloomFilter(rs io.ReadSeeker, bloomFilterOffset int64) (aBloomFilter, error) {
	//index := blockIndex{}
	rs.Seek(bloomFilterOffset, io.SeekStart)

	//compressReader, err := zlib.NewReader(rs)
	compressReader := rs
	var err error

	var bloomFilter aBloomFilter
	var bloomBlock bloomFilterBlock
	err = binary.Read(compressReader, binary.LittleEndian, &bloomBlock)
	if err != nil {
		return bloomFilter, err
	}

	data := make([]uint64, bloomBlock.NBuckets)
	err = binary.Read(compressReader, binary.LittleEndian, data)
	if err != nil {
		return bloomFilter, err
	}

	bloomFilter = bloomBlock.createFilter()

	bloomFilter.data = data

	return bloomFilter, nil

}
