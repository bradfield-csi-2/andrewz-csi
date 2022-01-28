package sstable

import (
	"bradfield/csi/distributed/lsmt/bloomFilter"
	"compress/lzw"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sort"
	"unsafe"
)

//var targetFileSize uint64 = 2000000
var TargetBlockSize uint32 = 4096 //64000
var MaxStringSize int = 1 << 16
var MaxFileSize int = 1 << 31

const MAX_KEYLEN = 4000
const MAX_VALLEN = 64000 //should i make it bigger?? --> maybe closer to 1MB

const MAX_UINT32 = 4294967295

//will make the 4 as one byte be special tombstone character
type Item struct {
	Key   string
	Value []byte
}

// Given a sorted list of key/value pairs, write them out according to the format you designed.
func Build(path string, sortedItems []Item) error {

	f, err := os.Create(path)

	if err != nil {
		return err
	}

	defer f.Close()

	n := len(sortedItems)
	filter := bloomFilter.NewBloomFilter(uint64(n), 10)
	lzwriter := lzw.NewWriter(f, lzw.LSB, 8)

	idxRecs := make([]IndexRecord, 0)
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("No Items to write")
	}
	rec := IndexRecord{sortedItems[0].Key, uint32(offset), 0}
	acc := 0
	blockItems := 0
	for _, item := range sortedItems {
		if acc > int(TargetBlockSize) {
			//start a new block
			rec.NItems = uint32(blockItems)
			idxRecs = append(idxRecs, rec)
			acc = 0
			blockItems = 0
			lzwriter.Close()
			lzwriter = lzw.NewWriter(f, lzw.LSB, 8)
			offset, err = f.Seek(0, io.SeekCurrent)
			if offset > MAX_UINT32 {
				return errors.New("offset is greater than max uint32")
			}
			if err != nil {
				return err
			}
			rec = IndexRecord{item.Key, uint32(offset), 0}
		}
		err = writeItem(lzwriter, item)
		filter.Add(item.Key)
		if err != nil {
			return err
		}
		rec.NItems++
		acc += len(item.Key) + len(item.Value) + 8 //8 bytes for two uint32
		blockItems++

	}
	rec.NItems = uint32(blockItems)
	idxRecs = append(idxRecs, rec)
	lzwriter.Close()
	lzwriter = lzw.NewWriter(f, lzw.LSB, 8)
	idxOffset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	writeIndex(lzwriter, idxRecs)
	lzwriter.Close()

	filterOffset, err := f.Seek(0, io.SeekCurrent)
	filter.Write(f)

	footer := TableFooter{idxOffset, filterOffset}
	_, err = f.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	err = binary.Write(f, binary.LittleEndian, footer)
	if err != nil {
		return err
	}

	return nil
}

func writeItem(w io.Writer, item Item) error {
	err := binary.Write(w, binary.LittleEndian, uint32(len(item.Key)))
	if err != nil {
		return err
	}
	binary.Write(w, binary.LittleEndian, *(*[]byte)(unsafe.Pointer(&item.Key)))
	if err != nil {
		return err
	}
	binary.Write(w, binary.LittleEndian, uint32(len(item.Value)))
	if err != nil {
		return err
	}
	binary.Write(w, binary.LittleEndian, *(*[]byte)(unsafe.Pointer(&item.Value)))
	return err
}

// A Table provides efficient access into sorted key/value data that's organized according
// to the format you designed.
//
// Although a Table shouldn't keep all the key/value data in memory, it should contain
// some metadata to help with efficient access (e.g. size, index, optional Bloom filter).
type Table struct {
	file   *os.File
	path   string
	index  *Index
	filter bloomFilter.BloomFilter
	cache  *blockCache
}

type TableFooter struct {
	IndexOffset       int64
	BloomFilterOffset int64
}

// Prepares a Table for efficient access. This will likely involve reading some metadata
// in order to populate the fields of the Table struct.
func LoadTable(path string, cacheSize int) (*Table, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	table := Table{}
	table.path = path

	if err != nil {
		return &table, err
	}
	var footer TableFooter
	footerOffset := 0 - int64(unsafe.Sizeof(footer))
	_, err = f.Seek(footerOffset, io.SeekEnd)

	if err != nil {
		return nil, err
	}

	err = binary.Read(f, binary.LittleEndian, &footer)
	if err != nil {
		return nil, err
	}
	table.file = f

	table.index, err = loadIndex(f, footer.IndexOffset)

	if err != nil {
		return &table, err
	}
	table.filter, err = bloomFilter.LoadBloomFilter(f, footer.BloomFilterOffset)
	table.cache = newBlockCache(cacheSize)

	return &table, err

}

func readItem(r io.Reader) (Item, error) {
	var keyLen uint32
	var valLen uint32
	var item Item
	err := binary.Read(r, binary.LittleEndian, &keyLen)
	if err != nil {
		return item, err
	}
	keybuf := make([]byte, keyLen)
	err = binary.Read(r, binary.LittleEndian, keybuf)
	if err != nil {
		return item, err
	}
	item.Key = string(keybuf)
	err = binary.Read(r, binary.LittleEndian, &valLen)
	if err != nil {
		return item, err
	}
	valbuf := make([]byte, valLen)
	err = binary.Read(r, binary.LittleEndian, valbuf)
	item.Value = valbuf
	return item, err
}

func (t *Table) Get(key string) ([]byte, bool, error) {
	if !t.filter.MaybeContains(key) {
		return []byte{}, false, nil
	}
	i := t.index.getLastLE(key)
	if i < 0 {
		return []byte{}, false, nil
	}
	items, ok := t.cache.Get(i)
	if !ok {
		items = t.getBlockItems(i)
		go t.cache.Cache(i, items)
	}

	j := sliceFirstGE(items, key)
	if j == len(items) {
		return []byte{}, false, nil
	}
	if items[j].Key == key {
		return items[j].Value, true, nil
	}
	return []byte{}, false, nil
}

func (t *Table) getBlockItems(blockIdx int) []Item {
	items := make([]Item, t.index.Records[blockIdx].NItems)
	_, err := t.file.Seek(int64(t.index.Records[blockIdx].BlockStart), io.SeekStart)
	if err != nil {
		panic(err)
	}
	lzreader := lzw.NewReader(t.file, lzw.LSB, 8)
	defer lzreader.Close()

	for i := 0; i < int(t.index.Records[blockIdx].NItems); i++ {
		item, err := readItem(lzreader)
		if err != nil {
			break
		}
		items[i] = item
	}
	return items
}

func (t *Table) CloseFile() error {
	return t.file.Close()
}

func (t *Table) OpenFile() error {
	var err error
	t.file, err = os.Open(t.path)
	return err
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
	i := t.index.getLastLE(startKey)
	if i < 0 {
		i = 0
	}
	var iter tableIterator
	iter.nBlocks = len(t.index.Records)
	iter.blockItems = t.getBlockItems(i)
	j := sliceFirstGE(iter.blockItems, startKey)
	if j == len(iter.blockItems) {
		iter.blockNum = i + 1
		iter.recNum = 0
		if i < iter.nBlocks {
			iter.blockItems = t.getBlockItems(i + 1)
		}
	} else {
		iter.blockNum = i
		iter.recNum = j
	}
	iter.tab = t
	iter.startKey = startKey
	iter.endKey = endKey
	iter.nItems = len(iter.blockItems)

	iter.item = iter.blockItems[iter.recNum]

	return &iter, nil

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
	nBlocks          int
	blockNum         int
	recNum           int
	nItems           int
	item             Item
	startKey, endKey string
}

func (it *tableIterator) Next() {
	nItems := len(it.blockItems)
	it.recNum++
	if it.recNum == nItems {
		it.blockNum++
		if it.blockNum == it.nBlocks {
			return
		}
		items := it.tab.getBlockItems(it.blockNum)
		it.blockItems = items
		it.recNum = 0
	}
	if it.blockNum < it.nBlocks {
		it.item = it.blockItems[it.recNum]
	}
}

func (it *tableIterator) Valid() bool {
	return it.blockNum < it.nBlocks && it.item.Key <= it.endKey
}
func (it *tableIterator) Item() Item {
	return it.item
}
