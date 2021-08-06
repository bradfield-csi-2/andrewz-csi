package table

import (
	//"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
  "time"
	//"compress/gzip"
)

type block struct {
	header  blockHeader
	structs []kvStruct
	bytes   []byte
}

type blockHeader struct {
	NItems        uint32
	ByteLen       uint32
	KvBytesOffset uint32
	//Size          uint32
}

type kvStruct struct {
	KeyOffset uint32
	KeyLen    uint16
	ValLen    uint16
}


type blockCache struct {
  num uint8
  cache [10]cachedBlock
}

type cachedBlock struct {
  start uint32
  //use maybe uint32 to save space
  lastUsed time.Time
  items []Item
}


func newBlockCache() blockCache {
  return blockCache{}
}

func (c *blockCache)Get(start uint32) ([]Item, bool) {
  for i := uint8(0); i < c.num; i++ {
    block := c.cache[i]
    if block.start == start {
      block.lastUsed = time.Now()
      return block.items, true
    }
  }
  var items []Item
  return items, false
}

func (c *blockCache)Cache(start uint32, items []Item) {
  if c.num == 10 {
    minTime, minBlock := c.cache[0].lastUsed, 0
    for i, block := range c.cache {
      if block.lastUsed.Before( minTime) {
        minTime = block.lastUsed
        minBlock = i
      }
    }

    c.cache[minBlock].start = start
    c.cache[minBlock].items = items
    c.cache[minBlock].lastUsed = time.Now()
  } else {
    c.cache[c.num].start = start
    c.cache[c.num].items = items
    c.cache[c.num].lastUsed = time.Now()
    c.num++
  }
}
 

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
