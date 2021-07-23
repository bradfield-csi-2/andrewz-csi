package table


import (
  "unsafe"
  "fmt"
  "os"
  "encoding/binary"
  "compress/zlib"
  //"compress/gzip"
)

var targetFileSize uint64 = 2000000
var blockOffset uint64 = 16 //index offset uint64 8 bytes + file size uint64 8 bytes
var blockHeaderSize uint64 = uint64(unsafe.SizeOf(blockHeader)) // kvbytes offset, block size
var MaxStringSize int = 1 < 16

type Item struct {
	Key, Value string
}

// Given a sorted list of key/value pairs, write them out according to the format you designed.
func Build(path string, sortedItems []Item) error {

  f, err := os.Create(path)

  if err != nil {
    return err
  }

  defer f.Close()

  buf := make([]byte, 50)

 // b := block{blockHeaderSize,make([]blockItem,0),make([]byte,0)}
//  byteIdx := uint64(0)
  //blockOffset := uint64(0)
  fileSize := uint64(blockOffset)

  _, err = f.Write(buf[:blockOffset])

  idxStructs := make([]indexBlockStruct,0,256)
  idxKeyBytes := make([]byte,0,1024)


  //TODO: Append blocks
  currFileOffset := int64(0)
  itemIter := newItemIterator(items)
  for newBlockItemRange, newBlockSize, eof := itemIter.getNewBlockItems(); !eof; newBlockItemRange, newBlockSize, eof = itemIter.getNewBlockItems() {
    //TODO: create and write new block and get block range and add index item
    newBlock := newBlock(newBlockItemRange, newBlockSize)
    newBlock.Write()
    nextFileOffset, err := f.Seek(0,io.SeekCurrent)
    if err != nil {
      return err
    }
    //TODO: create index block using item range for keys and fileOffsets for block size
    if nextFileOffset > MaxFileSize {
      return errors.New(fmt.Sprintf("file has grown larger than MaxFileSize: %d \n",MaxFileSize))
    }

    firstKey, lastKey := getStartEndKey(newBlockItemRange)
    indexKeyOffset := len(idxKeyBytes)
    newIndexBlockStruct := indexBlockStruct{
      uint32(indexKeyOffset),
      uint16(len(firstKey)),
      uint16(len(lastKey)),
      uint32(currFileOffset),
      uint32(nexFileOffset),
    }
    idxStructs = append(idxStructs, newIndexBlockStruct)
    idxKeyBytes = append(idxKeyBytes, firstKey...)
    idxKeyBytes = append(idxKeyBytes, lastKey...)

    currFileOffset = nextFileOffset
  }
  //TODO: create and append index block
  //TODO: update file/table header
 

	return nil
}

// A Table provides efficient access into sorted key/value data that's organized according
// to the format you designed.
//
// Although a Table shouldn't keep all the key/value data in memory, it should contain
// some metadata to help with efficient access (e.g. size, index, optional Bloom filter).
type Table struct {
	// TODO
  //header: file size (without footer or header? just blocks
  //index: two sets of byte arrays: block offset start key and end key offset, len, and k/v pairs
  //index starts with?? or just use slices
  size uint32
  index tableIndex
}

type tableHeader struct {
  indexOffset uint64
  size uint64
}

//table options struct?
//should have a block cache? - later

// Prepares a Table for efficient access. This will likely involve reading some metadata
// in order to populate the fields of the Table struct.
func LoadTable(path string) (*Table, error) {
  //read table header
  //then fetch index and populate
  f, err := os.Open(path)
  if err != nil {
    return nil, err
  }

  buf := make([]byte,16)
  _, err = f.Read(buf)
  var indexOffset,fileSize uint64
  indexOffset, err = binary.Read(buf)
  if err != nil {
    return nil, err
  }

  fileSize, err = binary.Read(buf[8:])
  if err != nil {
    return nil, err
  }


	return nil, nil
}

func (t *Table) Get(key string) (string, bool, error) {
  //search index
  //get block
  //search block?
	return "", false, nil
}

func (t *Table) RangeScan(startKey, endKey string) (Iterator, error) {
  //use index to get blocks
  //build iterator which does the work??
  //load a block at a time?
  //fetch next block in background?
	return nil, nil
}

type Iterator interface {
	// Advances to the next item in the range. Assumes Valid() == true.
	Next()

	// Indicates whether the iterator is currently pointing to a valid item.
	Valid() bool

	// Returns the Item the iterator is currently pointing to. Assumes Valid() == true.
	Item() Item
}
//#Index

type indexBlock struct {
  indexBlockHeader
  structs []indexBlockStruct
  bytes []byte
}

type indexBlockHeader struct {
  byteOffset uint32
  size uint32
}

type indexPage struct {
  startKey string
  endKey string
  blockStart uint32
  blockEnd uint32
}

type indexBlockStruct struct {
  startKeyOffset uint32
  startKeyLen uint16
  endKeyLen uint16
  //endKeyOffset uint32
  //keyLen uint32
  blockOffset uint32//max 4GB file -- need 64B in future?//or if aligned can chunk at 4KB blocks
  nextBlockOffset uint32
}


type indexPageKeys struct {
  start string
  end string
}


func newIndexBlock(structs []indexBlockStructs, bytes []byte, totKeyLen int) *indexBlock {
  hdr := indexBlockHeader{}
  byteOffset := len(structs) * int(unsafe.Sizeof(structs[0])) + totKeyLen
  size := int(unsafe.Sizeof(hdr)) + byteOffset
  hdr.byteOffset = uint32(byteOffset)
  hdr.size = uint32(size)
  return &indexBlock{hdr,structs,bytes}
}

func getStartEndKey(items []Item) (first, last, string) {// (keyPair indexPageKeys, totLen int) {
  ilen := len(items)
  first := items[0]
  last := items[ilen - 1]
  //totLen := len(first) + len(last)
  //return indexPageKeys{first.Key, last.Key}
  return first, last
}

func newIndex(keyPairs []indexPageKeys, totKeyLen int) tableIndex {
  //header fill and return
  return tableIndex{}
}

func (tIdx *tableIdx) bst(key string)  (uint64, bool) {
  //itemsLen := len(tIdx.indexItems)
  //if itemsLen < 1 {
    //panic("no items in index")
  //}
  //lastIdx := itemsLen - 1
  for i, page := range tIdx.indexItems {
    //sKeyOff := page.startKeyOffset
    //sKeyLen := page.endKeyOffset -1sKeyOff
    //eKeyOff := page.endKeyOffset
    //eKeyLen := page.keyLen - sKeyLen

    sKey := string(tIdx.keyBytes[page.startKeyOffset:page.endKeyOffset])
    eKey := string(tIdx.keyBytes[page.endKeyOffset:page.startKeyOffset+page.keyLen])

    if key < sKey {
      return 0, false
    }
    if key <= eKey {
      return page.blockoffset, true
    }
  }
  return 0, false

}


//#Block
//
type block struct {
  //size uint64
  header blockHeader
  structs []kvStruct
  bytes []byte
}

type blockHeader struct {
  kvBytesOffset uint32
  size uint32
}

func newBlock(items []Item, size uint32) *block {
  iLen := len(items)
  kvStructs := make([]kvStruct,iLen)

  newBlkHdr := blockHeader{}
  dataOffset := uint32(unsafe.Sizeof(newBlkHdr))
  kvBytesOffset := dataOffset + (uint32(unsafe.Sizeof(kvStructs[0])) * iLen)
  newBlkHdr.kvBytesOffset = kvBytesOffset
  newBlkHdr.size = size

  kvBytes := make([]byte,size - kvBytesOffset)
  
  byteOffset := uint32(0)
  for i,item := range items {
    keyLen := len(item.Key)
    valLen := len(item.Value)
    valOffset := byteOffset + uint32(valLen)
    kvLen := uint32(keyLen + valLen)
    newKVStruct :=  kvStruct{byteOffset,uint16(keyLen),uint16(valLen)}
    nextOffset = byteOffset + kvLen

    copy(kvBytes[byteOffset:valOffset],item.Key)
    copy(kvBytes[valOffset:nextOffset],item.Value)

    kvStructs[i] = newKVStruct

    byteOffset += nextOffset
  }


  return &block{newBlkHdr,kvStructs,kvBytes}

}

func (b *block) Write (w io.Writer) error {
  //TODO: write header and offsets and len and keys 
  //should compress first
  compressWriter, err := zlib.NewWriterLevel(w,zlib.BestCompression)
  //compressWriter, err := gzip.NewWriterLevel(w,gzip.BestCompression)

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

  err = compressWriter.Flush()
  if err != nil {
    return err
  }
 
}

type itemIterator struct {
  items []Items
  current int
}

func newItemIterator(items []Items) *itemIterator {
  return &itemIterator{items,0}
}

func (it *itemIterator)getNewBlockItems() (newBlockItems []Items, size uint32, eof bool) {
  iLen := len(it.items)
  size = 0
  blockItemStructSize := uint32(unsafe.SizeOf(blockItem))

  var j int

  for j = it.current ;j < iLen ; j++ {
    currItem := i.items[j]
    keyLen := len(currItem.Key)
    valLen := len(currItem.Value)
    if keyLen > MaxStringSize || valLen > MaxStringSize {
      return errors.New(fmt.Sprintf("string value is greater than MaxStringSize: %d\n",MaxStringSize))
    }
  
    size += blockItemStructSize + uint32(keyLen) + uint32(valLen)
    if size > targetBlockSize {
      break
    }
  }
  start = it.current
  it.current = j

  eof = false
  newBlockItems = it.items[start:j]

  if j == iLen {
    eof = true
  }
  
  return newBlockItems, size, eof
}




func newBlock() *block {
  return &block{size:blockHeaderSize}
}

type kvStruct struct {
  keyOffset uint32
  keyLen uint16
  valLen uint16
}


