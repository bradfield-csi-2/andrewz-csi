package main

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"fmt"
	math_rand "math/rand"
)

const MaxLevel int = 32 //16
const P float64 = 0.25  //.25

var myrand math_rand.Rand

type skipListOC struct {
	head  *slNode
	tail  *slNode
	level int
}

type slNode struct {
	item    Item
	forward []*slNode
}

func init() {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		panic("cannot seed math/rand package with cryptographically secure random number generator")
	}
	fmt.Println(b)
	math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

func randomLevel() int {
	lvl := 0
	for math_rand.Float64() < P && lvl < MaxLevel-1 {
		lvl++
	}
	return lvl
}

func newSkipListOC() *skipListOC {
	tail := &slNode{}
	hdFwd := make([]*slNode, MaxLevel)
	for i, _ := range hdFwd {
		hdFwd[i] = tail
	}
	head := &slNode{forward: hdFwd}
	return &skipListOC{head: head, tail: tail, level: 0}
}

func newSLNode(lvl int, key, val string) *slNode {
	fwd := make([]*slNode, lvl+1)
	return &slNode{
		item:    Item{key, val},
		forward: fwd,
	}
}

func (o *skipListOC) firstGE(key string) (*slNode, []*slNode) {
	update := make([]*slNode, MaxLevel)
	node := o.head //.forward[o.level]
	for i := o.level; i >= 0; i-- {
		for node.forward[i] != o.tail && node.forward[i].item.Key < key {
			node = node.forward[i]
		}
		update[i] = node
	}
	node = node.forward[0]
	return node, update
}

func (o *skipListOC) Get(key string) (string, bool) {
	node, _ := o.firstGE(key)
	if node == o.tail || node.item.Key != key {
		return "", false
	}
	return node.item.Value, true
}

func (o *skipListOC) Put(key, value string) bool {
	node, update := o.firstGE(key)
	if node.item.Key == key {
		node.item.Value = value
		return false
	} else {
		lvl := randomLevel()
		if lvl > o.level {
			for i := o.level + 1; i <= lvl; i++ {
				update[i] = o.head
			}
			o.level = lvl
		}
		newNode := newSLNode(lvl, key, value)
		for i := 0; i <= lvl; i++ {
			newNode.forward[i] = update[i].forward[i]
			update[i].forward[i] = newNode
		}
	}
	return true
}

func (o *skipListOC) Delete(key string) bool {
	node, update := o.firstGE(key)
	if node != o.tail && node.item.Key == key {
		for i := 0; i <= o.level; i++ {
			if update[i].forward[i] != node {
				break
			}
			update[i].forward[i] = node.forward[i]
		}
		for o.level > 0 && o.head.forward[o.level] == o.tail {
			o.level--
		}
		return true
	}
	return false
}

func (o *skipListOC) RangeScan(startKey, endKey string) Iterator {
	node, _ := o.firstGE(startKey)
	return &skipListOCIterator{o, node, startKey, endKey}
}

type skipListOCIterator struct {
	o                *skipListOC
	node             *slNode
	startKey, endKey string
}

func (iter *skipListOCIterator) Next() {
	iter.node = iter.node.forward[0]
}

func (iter *skipListOCIterator) Valid() bool {
	return iter.node != iter.o.tail && iter.node.item.Key <= iter.endKey
}

func (iter *skipListOCIterator) Key() string {
	return iter.node.item.Key
}

func (iter *skipListOCIterator) Value() string {
	return iter.node.item.Value
}
