package main

import (
	"encoding/binary"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const BUFPAGESIZE = 16384
const BUFPAGESHIFT = 14
const MAX_FILE_DESCRIPTORS = 20
const REC_PTR_START = 6

const (
	NORMAL = iota
	DELETED
	UPDATED
)

var DEBUG bool = false
var fileTokens chan struct{} = make(chan struct{}, MAX_FILE_DESCRIPTORS)

var BufMgr bufMgr

var TblMgr tblMgr

// create file descriptor semaphore??
//or just rely on os
//write method needs a file and offset
//keep track of open tables.. limit open tables/file descriptors

type tableBlock struct {
	table   string
	blockId int64
}

type bufMgr struct {
	running  bool
	nPages   int
	bufpages []bufpage
	//mempages []bufMemPage
	cache map[tableBlock]int
	//message channel?
	pageReqChan chan bufPageReq
}

type bufPageReq struct {
	table   string
	blockId int64
	retChan chan *bufpage
}

//Get

func initBufMgr(n int) bufMgr {
	m := make([]bufMemPage, n)
	b := make([]bufpage, n)
	for i := 0; i < n; i++ {
		b[i].data = m[i][:]
	}
	c := make(map[tableBlock]int)
	ch := make(chan bufPageReq)
	return bufMgr{false, n, b, c, ch}

}

func (b *bufMgr) run() {
	if b.running {
		panic("should not be running yet")
	}
	b.running = true
	go func() {
		for {
			req, ok := <-b.pageReqChan
			if ok {
				//TODO: process request
				b.processPageReq(req)
			} else {
				//Closed
				b.running = false
				return
			}
		}

	}()

}

func (b *bufMgr) stop() {
	if !b.running {
		panic("Should not stop buf mgr if not running")
	}
	close(b.pageReqChan)
	b.running = false
}

func (b *bufMgr) GetBufPage(table string, blockId int64) *bufpage {
	if !b.running {
		panic("Buf Manager is not running")
	}
	bufpgChan := make(chan *bufpage)
	req := bufPageReq{table, blockId, bufpgChan}
	b.pageReqChan <- req
	pg := <-bufpgChan
	return pg
}

func (b *bufMgr) processPageReq(req bufPageReq) {
	tb := tableBlock{req.table, req.blockId}
	if pgIdx, ok := b.cache[tb]; ok {
		atomic.AddInt32(&b.bufpages[pgIdx].pinCnt, 1)
		req.retChan <- &(b.bufpages[pgIdx])
		close(req.retChan)
		return
	}
	req.retChan <- b.evictAndGetBufpage(req)
	close(req.retChan)
}

//consider randomized and clock for better cache performance
//also consider prefetching and keeping pin count in batches
//WARNING: Cannot write a block more than one greater than file
//otherwise blank blocks with no metadata will exist
//need to protect / detect this eventually
func (b *bufMgr) evictAndGetBufpage(req bufPageReq) *bufpage {
	start := rand.Intn(b.nPages)
	for {
		for i := 0; i < b.nPages; i++ {
			curr := (start + i) % b.nPages
			if b.bufpages[curr].pinCnt == 0 {
				if b.bufpages[curr].clock {
					b.bufpages[curr].clock = false
					continue
				}
				tb := tableBlock{b.bufpages[curr].table, b.bufpages[curr].blockId}
				delete(b.cache, tb)
				tb.table = req.table
				tb.blockId = req.blockId
				b.cache[tb] = curr
				if blockExists(TblMgr.GetTableFilepath(req.table), req.blockId) {
					b.bufpages[curr].Load(req.table, TblMgr.GetTableFilepath(req.table), req.blockId)
				} else {
					b.bufpages[curr].Init(req.table, TblMgr.GetTableFilepath(req.table), req.blockId)
				}
				atomic.AddInt32(&b.bufpages[curr].pinCnt, 1)
				return &(b.bufpages[curr])
			}
		}
		time.Sleep(time.Second)
	}
}

func blockExists(filepath string, blockId int64) bool {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	info, err := f.Stat()

	if err != nil {
		panic(err)
	}

	blockStart := blockId << BUFPAGESHIFT

	if info.Size() > blockStart {
		return true
	}
	return false
}

type bufMemPage [BUFPAGESIZE]byte

type bufpage struct {
	pinCnt   int32
	filepath string
	table    string
	blockId  int64
	used     uint16
	upper    uint16
	lower    uint16
	rwmutex  sync.RWMutex
	clock    bool
	data     []byte
}

func (b *bufpage) Flush() {
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()
	offset := b.blockId << BUFPAGESHIFT // * 4096
	f, err := os.OpenFile(b.filepath, os.O_WRONLY, 0)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	_, err = f.WriteAt(b.data, int64(offset))
	if err != nil {
		panic(err)
	}
	err = f.Sync()
	if err != nil {
		panic(err)
	}
}

/*
func (b *bufpage) CheckAndGetPage(table string, blockId int) bool {
	b.rwmutex.RLock()
	defer b.rwmutex.RUnlock()
	if b.table == table && b.blockId == blockId {
		atomic.AddInt32(&b.pinCnt, 1)
		return true
	}
	return false
}
*/

func (b *bufpage) Load(table, filepath string, blockId int64) {
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()
	b.table = table
	b.blockId = blockId
	b.filepath = filepath
	offset := b.blockId << BUFPAGESHIFT // * 4096
	f, err := os.Open(b.filepath)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	n, err := f.ReadAt(b.data, int64(offset))
	if n != BUFPAGESIZE {
		panic("weird number read")
	}
	if err != nil {
		panic(err)
	}

	b.used = binary.LittleEndian.Uint16(b.data[:])
	b.upper = binary.LittleEndian.Uint16(b.data[2:])
	b.lower = binary.LittleEndian.Uint16(b.data[4:])
}

func (b *bufpage) Init(table, filepath string, blockId int64) {
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()
	b.table = table
	b.blockId = blockId
	b.filepath = filepath   //TODO: Get filename from table manager
	b.used = REC_PTR_START  //binary.LittleEndian.Uint16(b.data[:])
	b.upper = REC_PTR_START //binary.LittleEndian.Uint16(b.data[2:])
	b.lower = BUFPAGESIZE   //binary.LittleEndian.Uint16(b.data[4:])

	binary.LittleEndian.PutUint16(b.data[:], b.used)
	binary.LittleEndian.PutUint16(b.data[2:], b.upper)
	binary.LittleEndian.PutUint16(b.data[4:], b.lower)

	for i := 6; i < BUFPAGESIZE; i++ {
		b.data[i] = 0
	}
}

func (b *bufpage) GetRecord(rowOffset int) []byte {
	//assuming we already now the record is here, since we know the rowOffset
	//wait should we know that..??
	// maybe we know first and last row ids, but would we know the offset??
	//maybe we should binary search based on row id??
	b.rwmutex.RLock()
	defer b.rwmutex.RUnlock()
	start := REC_PTR_START             // uint16 * 3
	recOff := start + (rowOffset << 2) //*2

	dataOff := binary.LittleEndian.Uint16(b.data[recOff:])
	dataLen := binary.LittleEndian.Uint16(b.data[recOff+2:])
	rowData := make([]byte, dataLen)
	copy(rowData, b.data[dataOff:])
	return rowData
}

func (b *bufpage) HasRecord(rowOffset int) bool {
	//assuming we already now the record is here, since we know the rowOffset
	//wait should we know that..??
	// maybe we know first and last row ids, but would we know the offset??
	//maybe we should binary search based on row id??
	b.rwmutex.RLock()
	defer b.rwmutex.RUnlock()
	start := REC_PTR_START             // uint16 * 3: TODO: make this a constant Or use some math
	recOff := start + (rowOffset << 2) //*2 TODO: use size
	return uint16(recOff) < b.upper
}

func (b *bufpage) CanWriteRecord(recLen uint16) bool {
	b.rwmutex.RLock()
	defer b.rwmutex.RUnlock()

	if b.upper+4 > b.lower-recLen { //TODO: conststant increment
		return false
	}
	return true
}

//TODO: use constants rather than inline hardcode values
//returns num of slots vacuumed
func (b *bufpage) Vacuum() int {
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()
	cnt := 0
	var bytesRemoved uint16 = 0
	var recFreePtr uint16 = REC_PTR_START
	var dataFreePtr uint16 = BUFPAGESIZE - 1
	var recPtr uint16 = REC_PTR_START
	//Todo: start at beginning
	//check record
	for recPtr < b.upper {
		dataOff := binary.LittleEndian.Uint16(b.data[recPtr:])
		dataLen := binary.LittleEndian.Uint16(b.data[recPtr+2:]) //TODO: const uin16 size
		firstByte := b.data[dataOff]
		if firstByte == NORMAL {
			for srcPtr := dataOff + dataLen - 1; srcPtr >= dataOff; srcPtr-- {
				b.data[dataFreePtr] = b.data[srcPtr]
				dataFreePtr--
			}
			b.data[recFreePtr] = b.data[recPtr]
			b.data[recFreePtr+1] = b.data[recPtr+1]
			b.data[recFreePtr+2] = b.data[recPtr+2]
			b.data[recFreePtr+3] = b.data[recPtr+3]
			recFreePtr += 4
		} else {
			bytesRemoved += 4 + dataLen
			cnt++
		}
		recPtr += 4
	}
	b.used -= bytesRemoved
	b.upper = recFreePtr
	b.lower = dataFreePtr

	binary.LittleEndian.PutUint16(b.data[:], b.used)
	binary.LittleEndian.PutUint16(b.data[2:], b.upper)
	binary.LittleEndian.PutUint16(b.data[4:], b.lower)

	return cnt
}

func (b *bufpage) DeleteRecord(rowOffset int) bool {
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()
	start := REC_PTR_START             // uint16 * 3
	recOff := start + (rowOffset << 2) //*2

	dataOff := binary.LittleEndian.Uint16(b.data[recOff:])
	dataLen := binary.LittleEndian.Uint16(b.data[recOff+2:]) //TODO: constant size of uint16
	if dataLen < 1 {
		panic("should not be a record this small")
	}
	b.data[dataOff] |= DELETED
	return true
}

func (b *bufpage) UpdateRecord(rowOffset int, newRec record) bool {
	recLen := uint16(len(newRec))
	if !b.CanWriteRecord(recLen) {
		return false
		//TODO: write on next page
	}
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()
	start := REC_PTR_START             // uint16 * 3
	recOff := start + (rowOffset << 2) //*2

	dataOff := binary.LittleEndian.Uint16(b.data[recOff:])
	dataLen := binary.LittleEndian.Uint16(b.data[recOff+2:]) //TODO: constant size of uint16
	if dataLen < 1 {
		panic("should not be a record this small")
	}
	b.data[dataOff] |= UPDATED
	//rowData := make([]byte, dataLen)
	//copy(rowData, b.data[dataOff:])
	//return rowData
	//WARNING: must make sure can write record

	binary.LittleEndian.PutUint16(b.data[b.upper:], b.lower-recLen)
	binary.LittleEndian.PutUint16(b.data[b.upper+2:], recLen)
	b.upper += 4 //TODO: constant size of recPtr
	binary.LittleEndian.PutUint16(b.data[2:], b.upper)
	copy(b.data[b.lower-recLen:], newRec)

	//TODO: still using this convention because there is not space until
	//vacuuming process completes. need to tracking space to clean up tho
	//so vacuuming knows when to start
	b.used += 4 + recLen //TODO: constant size of rec Ptr
	binary.LittleEndian.PutUint16(b.data[:], b.used)

	if b.lower < recLen {
		panic("weird values")
	}
	b.lower -= recLen
	binary.LittleEndian.PutUint16(b.data[4:], b.lower)
	return true

}

func (b *bufpage) WriteRecord(rec record) bool {
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()
	recLen := uint16(len(rec))

	if b.upper+4 > b.lower-recLen {
		return false
	}
	binary.LittleEndian.PutUint16(b.data[b.upper:], b.lower-recLen)
	binary.LittleEndian.PutUint16(b.data[b.upper+2:], recLen)
	b.upper += 4
	binary.LittleEndian.PutUint16(b.data[2:], b.upper)
	copy(b.data[b.lower-recLen:], rec)

	b.used += 4 + recLen
	binary.LittleEndian.PutUint16(b.data[:], b.used)

	if b.lower < recLen {
		panic("weird values")
	}
	b.lower -= recLen
	binary.LittleEndian.PutUint16(b.data[4:], b.lower)
	return true

}

func (b *bufpage) Release() {
	b.clock = true
	atomic.AddInt32(&b.pinCnt, -1)
}

type recPtr struct {
	offset uint16
	len    uint16
}

/*
layout
row = offset | len

row = some data type
string = uint16 | bytes
other data types pretty self explantory

schema lives outside of file with table info

table catalog stores file name - where does block data go?
and file_name of table meta data- schema, check,
in index?
size of table file
pointer to index maybe?

how to implement reader writer operations??


*/
