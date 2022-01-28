package sstable

import (
	//"compress/zlib"

	"math/rand"
	//"sync"
	"sync/atomic"
	//"compress/gzip"
)

/*
type block struct {
	header  blockHeader
	structs []kvStruct
	bytes   []byte
}
*/
type block []byte

type blockHeader struct {
	//NItems        uint32
	//ByteLen       uint32
	//KvBytesOffset uint32
	//Size          uint32
	RecStart uint32
	RecEnd   uint32
}

type kvStruct struct {
	KeyOffset uint32
	KeyLen    uint16
	ValLen    uint16
}

type blockCache struct {
	num   int
	cache []*cachedBlock
	index map[int]int
}

type cachedBlock struct {
	blockNum int
	pinCnt   int32
	marked   bool
	clock    bool
	//use maybe uint32 to save space
	//lastUsed time.Time
	//lock  sync.RWMutex
	items []Item
}

func newBlockCache(cacheSize int) *blockCache {
	c := blockCache{0, make([]*cachedBlock, cacheSize), make(map[int]int)}
	for i := 0; i < cacheSize; i++ {
		c.cache[i] = &cachedBlock{blockNum: -1}
	}
	return &c
}

func (c *blockCache) Get(blockNum int) ([]Item, bool) {
	var items []Item
	cacheIdx, ok := c.index[blockNum]
	for ok {
		cblock := c.cache[cacheIdx]
		cnt := atomic.AddInt32(&cblock.pinCnt, 1)
		if (cnt > 1 && cblock.marked) || cblock.blockNum != blockNum { // || cblock.blockNum != blockNum
			atomic.AddInt32(&cblock.pinCnt, -1)
			cacheIdx, ok = c.index[blockNum]
			continue
		} else if cnt == 1 {
			cblock.marked = false
		}
		cblock.clock = true
		items = cblock.items
		atomic.AddInt32(&cblock.pinCnt, -1)
		break
	}
	return items, ok
}

func (c *blockCache) Cache(blockNum int, items []Item) {
	//idx, ok := c.index[blockNum]
	n := len(c.cache)
	rstart := rand.Intn(n)
	i := 0
	for {
		if _, ok := c.index[blockNum]; ok {
			return
		}
		idx := (rstart + i) % n
		if c.cache[idx].pinCnt == 0 {
			cblock := c.cache[idx]
			if cblock.clock {
				cblock.clock = false
				continue
			}
			cblock.marked = true
			cnt := atomic.AddInt32(&cblock.pinCnt, 1)
			if cnt > 1 {
				atomic.AddInt32(&cblock.pinCnt, -1)
				continue
			}
			delete(c.index, cblock.blockNum)
			c.index[blockNum] = idx
			cblock.items = items
			cblock.blockNum = blockNum
			cblock.marked = false
			atomic.AddInt32(&cblock.pinCnt, -1)
			return
		}
		i++
	}
}

/*
func loadBlockItems(rs io.ReadSeeker, page indexPage) ([]Item, error) {
	blkHdr := blockHeader{}

	rs.Seek(int64(page.blockStart), io.SeekStart)
	//compressReader, err := zlib.NewReader(rs)
	var err error
	compressReader := rs

	err = binary.Read(compressReader, binary.LittleEndian, &blkHdr)
	if err != nil {
		return nil, err
	}

	structs := make([]kvStruct, blkHdr.NItems)
	bytes := make([]byte, blkHdr.ByteLen)
	items := make([]Item, blkHdr.NItems)
	err = binary.Read(compressReader, binary.LittleEndian, structs)
	if err != nil {
		return items, err
	}

	err = binary.Read(compressReader, binary.LittleEndian, bytes)
	if err != nil {
		return items, err
	}

	strData := string(bytes)

	for i, kvStruct := range structs {
		keyOffset := int(kvStruct.KeyOffset)
		valOffset := keyOffset + int(kvStruct.KeyLen)
		end := valOffset + int(kvStruct.ValLen)
		items[i] = Item{strData[keyOffset:valOffset], strData[valOffset:end]}
	}

	return items, nil
}

func newBlock(items []Item, byteLen int, filter *aBloomFilter) *block {
	iLen := len(items)
	kvStructs := make([]kvStruct, iLen)

	newBlkHdr := blockHeader{}
	kvBytesOffset := uint32(unsafe.Sizeof(kvStructs[0])) * uint32(iLen)
	newBlkHdr.NItems = uint32(iLen)
	newBlkHdr.ByteLen = uint32(byteLen) //size - kvBytesOffset
	newBlkHdr.KvBytesOffset = kvBytesOffset

	kvBytes := make([]byte, byteLen)

	byteOffset := uint32(0)
	for i, item := range items {
		keyLen := len(item.Key)
		valLen := len(item.Value)
		valOffset := byteOffset + uint32(keyLen)
		kvLen := uint32(keyLen + valLen)
		newKVStruct := kvStruct{byteOffset, uint16(keyLen), uint16(valLen)}
		nextOffset := byteOffset + kvLen

		copy(kvBytes[byteOffset:valOffset], item.Key)
		copy(kvBytes[valOffset:nextOffset], item.Value)

		kvStructs[i] = newKVStruct

		byteOffset = nextOffset

		filter.add(item.Key)
	}

	return &block{newBlkHdr, kvStructs, kvBytes}

}

func (b *block) Write(w io.Writer) error {
	//compressWriter, err := zlib.NewWriterLevel(w, zlib.BestCompression)
	//defer compressWriter.Close()
	var err error
	compressWriter := w

	if err != nil {
		return err
	}

	err = binary.Write(compressWriter, binary.LittleEndian, b.header)
	if err != nil {
		return err
	}

	err = binary.Write(compressWriter, binary.LittleEndian, b.structs)
	if err != nil {
		return err
	}
	err = binary.Write(compressWriter, binary.LittleEndian, b.bytes)
	if err != nil {
		return err
	}

	//err = compressWriter.Flush()
	if err != nil {
		return err
	}

	return nil
}

type itemIterator struct {
	items   []Item
	current int
}

func newItemIterator(items []Item) *itemIterator {
	return &itemIterator{items, 0}
}

func (it *itemIterator) getNewBlockItems() (newBlockItems []Item, byteLen int, eof bool, err error) {
	iLen := len(it.items)
	size := uint32(0)
	kvStruct := kvStruct{}
	blockItemStructSize := uint32(unsafe.Sizeof(kvStruct))
	byteLen = 0

	var j int

	for j = it.current; j < iLen; j++ {
		currItem := it.items[j]
		keyLen := len(currItem.Key)
		valLen := len(currItem.Value)
		if keyLen > MaxStringSize || valLen > MaxStringSize {
			return it.items[0:0], 0, false, errors.New(fmt.Sprintf("string value is greater than MaxStringSize: %d\n", MaxStringSize))
		}

		size += blockItemStructSize + uint32(keyLen) + uint32(valLen)
		byteLen += keyLen + valLen
		if size > TargetBlockSize {
			j++
			break
		}
	}
	start := it.current
	it.current = j

	eof = false
	newBlockItems = it.items[start:j]

	if j == iLen {
		eof = true
	}

	return newBlockItems, byteLen, eof, nil
}
*/
