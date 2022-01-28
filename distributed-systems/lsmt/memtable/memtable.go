package memtable

//create a write ahead log that inserts records to eventually create a memtable
//how to handle concurrency?
//if created? --> just create?
//what if an item before was existing --> how to overwrite?
//just create and on compaction we will serialize and overwrite

//set if item exists then can overwrite --> but what about concurrency?
//if concurrent-> should create a new record and retire the old-->
//delete-> if deleting than just retire the old?

//what if deleting a record that doesn't exist? //can we create a tombstone
//that will be compacted at a later time?
//if we enter a tombstone for a non existing record doesn't it
//have to be tracked in the bloomfilter? kind of annoying
//only create if bloomfilter is true
//will have to find anyway because the bloomfilter makes us have to search it anyway
//above is downstream consideration?

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	math_rand "math/rand"
	"sync/atomic"
)

//const MaxLevel int = 32//16
//const P float64 = 0.25//.25

var myrand math_rand.Rand

type LogRecord struct {
	Flags   uint32 //set/delete record -> do I care about replaying values?
	Valid   uint64 //will it every wrap around?
	Expired uint64
	Key     string
	Value   []byte
}

//want to just overwrite last value, but in order of lsn
//don't care about concurrent writes really?
//do i care to commit transaction as durable first
//or can i read from memory and assume it will work out
//idk, don't really want to coordivate reads.
//what about read your writes
//1. one. when to ok a write
//after the commit to log
//or after inserted into skiplist
//unless allowing concurrent items to exist, but then need to deal with
//inserting in correct order
//commit to the log, then insert into the skip list, then ok
type Item struct {
	//flags uint16
	//marked bool
	//updating bool
	Flags uint16
	Key   string
	Value []byte
}

//if it is concurrent on same system i have the transaction manager
//to decide what order to play the logs in
//but if distributed how do i resolve conflicts?

//minimal concurrency -> locking on some rows?
//don't care about transactions and gets
//just want writes to be in proper order?
/*
type memtable struct {
}
*/

type MemTable interface {
	Get(key string) ([]byte, bool)
	Put(key string, lsn uint64) bool
	Delete(key string, lsn uint64) bool
	//RangeScan(startKey, endKey string) Iterator
	//Freeze() bool
	Write(filepath string) error
	//LOCK
}

//should I read everything into memory inorder to do this???
//or just reupdate when files are compacted/updated?
//use lock but becareful to always finish using range scan as quick as possible
//can maintain two copies of a level until it's done

type Iterator interface {
	// Advances to the next item in the range. Assumes Valid() == true.
	Next()

	// Indicates whether the iterator is currently pointing to a valid item.
	Valid() bool

	// Returns the Key for the item the iterator is currently pointing to.
	// Assumes Valid() == true.
	Key() string

	// Returns the Value for the item the iterator is currently pointing to.
	// Assumes Valid() == true.
	Value() string
}

//needs to be concurrent
//support tombstones
//damn does my sstable need to have tombstones??

type skipList struct {
	head     *slNode
	tail     *slNode
	level    int32
	maxLevel int32
	p        float64
}

type slNode struct {
	deleted bool
	valid   bool
	marked  uint32
	pinCnt  int32
	lsn     uint64
	item    Item
	forward []*slNode
}

func init() {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		panic("cannot seed math/rand package with cryptographically secure random number generator")
	}
	math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

func randomLevel(maxLevel int32, p float64) int32 {
	lvl := int32(0)
	for math_rand.Float64() < p && lvl < maxLevel-1 {
		lvl++
	}
	return lvl
}

func newSkipList(maxLevel int32, p float64) *skipList {
	tail := &slNode{}
	hdFwd := make([]*slNode, maxLevel)
	for i, _ := range hdFwd {
		hdFwd[i] = tail
	}
	head := &slNode{forward: hdFwd}
	return &skipList{head: head, tail: tail, level: 0, maxLevel: maxLevel, p: p}
}

func newSLNode(lvl int32, key string, value []byte) *slNode {
	fwd := make([]*slNode, lvl+1)
	return &slNode{
		item:    Item{0, key, value},
		forward: fwd,
	}
}

func (sl *skipList) firstGE(key string) (*slNode, []*slNode) {
	update := make([]*slNode, sl.maxLevel)
	node := sl.head //.forward[o.level]
	for i := sl.level; i >= 0; i-- {
		for node.forward[i] != sl.tail && node.forward[i].item.Key < key {
			node = node.forward[i]
		}
		update[i] = node
	}
	node = node.forward[0]
	return node, update
}

func (sl *skipList) firstGELock(key string) (*slNode, []*slNode) {
	update := make([]*slNode, sl.maxLevel)
	node := sl.head //.forward[o.level]
	for i := sl.level; i >= 0; i-- {
		for node.forward[i] != sl.tail && node.forward[i].item.Key < key {
			node = node.forward[i]
		}
		update[i] = node
	}
	node = node.forward[0]
	return node, update
}

func (sl *skipList) Get(key string) (val []byte, valid bool) {
	node, _ := sl.firstGE(key)
	if node == sl.tail || node.item.Key != key || !node.valid {
		return val, false
	}

	for {
		if node.marked == 1 {
			continue
		}
		atomic.AddInt32(&node.pinCnt, 1)
		if node.marked == 0 {
			break
		}
		atomic.AddInt32(&node.pinCnt, -1)
	}

	val = node.item.Value
	atomic.AddInt32(&node.pinCnt, -1)

	return val, true
}

//damn i need to do pretty much the same thing for delete
func (sl *skipList) put(key string, value []byte, lsn uint64, delete bool) bool {
	//wrap all in retry logic
	//var node *slNode
	//var update []*slNode
retry:
	node, update := sl.firstGE(key)
	if node.item.Key == key {
		//node.item.Value = value
		return node.atomicPut(key, value, lsn, delete)
	}
	//need to lock all update levels
	lvl := randomLevel(sl.maxLevel, sl.p)
	if lvl > sl.level {
		for i := sl.level + 1; i <= lvl; i++ {
			update[i] = sl.head
		}
	}
	if !sl.acquireUpdateLocks(update, key, lvl) {
		goto retry
	}
	slLevel := sl.level
	for lvl > slLevel {
		atomic.CompareAndSwapInt32(&sl.level, slLevel, lvl)
		slLevel = sl.level
	}
	newNode := newSLNode(lvl, key, value)

	newNode.lsn = lsn
	newNode.deleted = delete

	//unlock all update node
	insertNewNode(newNode, update, lvl)
	unlockUpdateNodes(update, lvl)

	return true
}

func insertNewNode(newNode *slNode, update []*slNode, lvl int32) {
	for i := lvl; i >= 0; i-- {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}
}

func unlockUpdateNodes(update []*slNode, lvl int32) {
	for i := lvl; i >= 0; i-- {
		swapped := atomic.CompareAndSwapUint32(&update[i].marked, 1, 0)
		if !swapped {
			panic("concurrency bug found while unlocking")
		}
	}
}

func (sl *skipList) acquireUpdateLocks(update []*slNode, key string, lvl int32) bool {
	for i := int32(0); i <= lvl; i++ {
		//can check if it's the same as the last node by one level down
		//otherwise it's a new node
		//newNode.forward[i] = update[i].forward[i]
		//update[i].forward[i] = newNode
		if i > 0 && update[i] == update[i-1] {
			continue
		}
		unode := update[i] //wait what about locks already acquired?
		var swapped bool
		for {
			swapped = atomic.CompareAndSwapUint32(&unode.marked, 0, 1)
			if !swapped {
				continue
			}
		}
		//acquired rights to this node
		//acquired marked access
		for unode.pinCnt != 0 {
			//spin
		}
		//acquired priveleges
		for unode.forward[i] != sl.tail && unode.forward[i].item.Key <= key {
			swapped = atomic.CompareAndSwapUint32(&unode.marked, 1, 0)
			if !swapped {
				panic("more concurrency bugs")
			}
			if unode.forward[i].item.Key == key {
				// can only happen if i == 0
				if i != 0 {
					panic("another concurrency bug in put")
				}
				return false //somehow was inserted by another process
			}
			unode = unode.forward[i]
			for {
				//acaquire rights to new node
				swapped = atomic.CompareAndSwapUint32(&unode.marked, 0, 1)
				if swapped {
					break
				}

			}
			continue
		}
		//acquired appropriate update node and
		update[i] = unode
	}
	return true
}

func (node *slNode) atomicPut(key string, value []byte, lsn uint64, delete bool) bool {
	for {
		//atomic.AddInt32(&node.pinCnt, 1)
		swapped := atomic.CompareAndSwapUint32(&node.marked, 0, 1)
		if !swapped {
			continue
		}
		for node.pinCnt != 0 {
			//loop and keep trying
		}
		break
	}
	//have exclusive write on the node now
	if lsn > node.lsn {
		node.item.Value = value
		node.lsn = lsn
		node.deleted = delete
	}
	swap := atomic.CompareAndSwapUint32(&node.marked, 1, 0)
	if !swap {
		panic("concurrency bug")
	}
	return false //??
}

func (sl *skipList) Put(key string, value []byte, lsn uint64) bool {
	return sl.put(key, value, lsn, false)
}

//need to fix this
func (sl *skipList) Delete(key string, lsn uint64) bool {
	return sl.put(key, []byte{}, lsn, true)
}
