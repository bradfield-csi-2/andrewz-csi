package table

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"unsafe"
	//"compress/gzip"
)

//var targetFileSize uint64 = 2000000
var TargetBlockSize uint32 = 4000 //64000
var MaxStringSize int = 1 << 16
var MaxFileSize int = 1 << 31

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

	idxStructs := make([]indexBlockStruct, 0, 256)
	idxKeyBytes := make([]byte, 0, 1024)

  filter := newBloomFilter(uint64(len(sortedItems)), 10)

	currFileOffset, nextFileOffset := int64(0), int64(0)
	itemIter := newItemIterator(sortedItems)
	newBlockItemRange, blockByteLen, eof, err := itemIter.getNewBlockItems()
	for {
		newBlock := newBlock(newBlockItemRange, blockByteLen, filter)
		err = newBlock.Write(f)
		if err != nil {
			return err
		}
		nextFileOffset, err = f.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		if nextFileOffset > int64(MaxFileSize) {
			return errors.New(fmt.Sprintf("file has grown larger than MaxFileSize: %d \n", MaxFileSize))
		}

		//firstKey, lastKey := getStartEndKey(newBlockItemRange)
		//firstKey := getStartKey(newBlockItemRange)
		lastKey := getEndKey(newBlockItemRange)
		indexKeyOffset := len(idxKeyBytes)
		newIndexBlockStruct := indexBlockStruct{
			uint32(indexKeyOffset),
			//uint16(len(firstKey)),
			uint16(len(lastKey)),
			uint32(currFileOffset),
			//uint32(nextFileOffset),
		}
		idxStructs = append(idxStructs, newIndexBlockStruct)
		//idxKeyBytes = append(idxKeyBytes, firstKey...)
		idxKeyBytes = append(idxKeyBytes, lastKey...)

		currFileOffset = nextFileOffset
		if err != nil {
			return err
		}
		if eof {
			break
		}
		newBlockItemRange, blockByteLen, eof, err = itemIter.getNewBlockItems()
	}
	if err != nil {
		return err
	}
	indexBlock := newIndexBlock(idxStructs, idxKeyBytes, len(idxKeyBytes))
  fmt.Println(len(idxStructs))
	err = indexBlock.Write(f)
	if err != nil {
		return err
	}

	nextFileOffset, err = f.Seek(0, io.SeekCurrent)

  if err != nil {
		return err
	}
	
  err = filter.Write(f)

	if err != nil {
		return err
	}
	
	footer := tableFooter{currFileOffset, nextFileOffset}
	err = binary.Write(f, binary.LittleEndian, footer)
	if err != nil {
		return err
	}
	return nil
}

// A Table provides efficient access into sorted key/value data that's organized according
// to the format you designed.
//
// Although a Table shouldn't keep all the key/value data in memory, it should contain
// some metadata to help with efficient access (e.g. size, index, optional Bloom filter).
type Table struct {
	file  *os.File
	index blockIndex
  filter aBloomFilter
  cache blockCache
}

type tableFooter struct {
	IndexOffset int64
	//IndexEnd    int64
  BloomFilterOffset int64
}

// Prepares a Table for efficient access. This will likely involve reading some metadata
// in order to populate the fields of the Table struct.
func LoadTable(path string) (*Table, error) {
	//read table header
	//then fetch index and populate
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	//var indexOffset int64
	var footer tableFooter
	footerOffset := int64(unsafe.Sizeof(footer)) * -1
	f.Seek(footerOffset, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	err = binary.Read(f, binary.LittleEndian, &footer)
	if err != nil {
		return nil, err
	}

	table := Table{file: f}
	table.index, err = loadIndex(f, footer.IndexOffset)
	if err != nil {
		return nil, err
	}

	table.filter, err = loadBloomFilter(f, footer.BloomFilterOffset)
	if err != nil {
		return nil, err
	}

  table.cache = newBlockCache()

	return &table, nil
}

func (t *Table) Get(key string) (string, bool, error) {
  maybeContains := t.filter.maybeContains(key)
  if !maybeContains {
    //fmt.Println("hitting")
		return "", false, nil
  }

	i := t.index.GetFirstGE(key)

	if i == len(t.index.pages) {
		return "", false, nil
	}

	page := t.index.pages[i]


  items, inCache := t.cache.Get(page.blockStart)

  if !inCache {
    var err error
	  items, err = loadBlockItems(t.file, page)
	  if err != nil {
		  return "", false, err
	  }
    t.cache.Cache(page.blockStart, items)
  }

	j := sliceFirstGE(items, key)
	if j == len(items) {
		return "", false, nil
	}
	return items[j].Value, true, nil
}

func sliceFirstGE(items []Item, key string) int {
	// Use the binary search implementation from the standard library
	return sort.Search(len(items), func(i int) bool {
		return items[i].Key >= key
	})

	// If we wanted to use linear search instead, we could do this:
	// i := 0
	// for i < len(items) && items[i].Key < key {
	//     i++
	// }
	// return i
}

func (t *Table) RangeScan(startKey, endKey string) (Iterator, error) {
	//use index to get blocks
	//build iterator which does the work??
	//load a block at a time?
	//fetch next block in background?
	i := t.index.GetFirstGE(startKey)

	if i == len(t.index.pages) {
		return &tableIterator{t, make([]Item, 0), len(t.index.pages), i, 0, Item{}, startKey, endKey}, nil
	}

	page := t.index.pages[i]

	items, err := loadBlockItems(t.file, page)
	if err != nil {
		return &tableIterator{}, err
	}

	j := sliceFirstGE(items, startKey)
	if j == len(items) {
		return &tableIterator{}, errors.New("something wrong with slice firstGE in range scan function")
	}
	return &tableIterator{t, items, len(t.index.pages), i, j, items[j], startKey, endKey}, nil
}

type Iterator interface {
	// Advances to the next item in the range. Assumes Valid() == true.
	Next()

	// Indicates whether the iterator is currently pointing to a valid item.
	Valid() bool

	// Returns the Item the iterator is currently pointing to. Assumes Valid() == true.
	Item() Item
}

type tableIterator struct {
	tab              *Table
	blockItems       []Item
	nPages           int
	pageNum          int
	recNum           int
	item             Item
	startKey, endKey string
}

func (it *tableIterator) Next() {
	nItems := len(it.blockItems)
	it.recNum++
	if it.recNum == nItems {
		it.pageNum++
		//load new block and items
		page := it.tab.index.pages[it.pageNum]
		items, err := loadBlockItems(it.tab.file, page)
		if err != nil {
			it.pageNum = it.nPages
		}
		it.blockItems = items
		it.recNum = 0
	}
	if it.pageNum < it.nPages {
		//set item
		it.item = it.blockItems[it.recNum]
	}
}

func (it *tableIterator) Valid() bool {
	return it.pageNum < it.nPages && it.item.Key <= it.endKey

}
func (it *tableIterator) Item() Item {
	return it.item
}
