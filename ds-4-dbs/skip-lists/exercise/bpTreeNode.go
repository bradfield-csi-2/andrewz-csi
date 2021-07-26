package main

import "sort"

const (
  InternalNode = iota
  LeafNode
)

type BpTreeNode interface {
  //Get key
  Get(key string) (string, bool)
  //Put
  Put(key, value string, pageSize int) (string, *Page, bool)
  //Delete
  Delete(key string) bool
  //GetPag
  GetFirstPage(key string) *Page
}


type BpInternalNode struct {
  entries []BPNodeEntry
}

type BPNodeEntry struct {
  key string
  node BPTreeNode
}


func (bpin *BpInternalNode)NodeType() int {
  return InternalNode
}

func bpNodeEntryFirstGE(entries []BPNodeEntry, key) int {
  return sort.Search(len(items), func(i int) bool {
    return entries[i].Key >= key
  })
}

func (bpin *BpInternalNode)Get(key string) (string, bool) {
  i := bpNodeEntryFirstGE(bpin.entries, key)
  if i != 0 {
    //return nil, false
    i--
  }
  return bpin.entries[i].node.Get(key)
}  

func (bpin *BpInternalNode)Put(key, value string, fanout int) bool {

}

/*
type BpTreeLeaf struct {
  entries []PageEntry
}

type PageEntry struct {
  key string
  page *Page
}


func (leaf *BPTreeLeaf)NodeType() int {
  return LeafNode
}

func pageEntryFirstGE(entries []PageEntry, key) int {
  return sort.Search(len(items), func(i int) bool {
    return entries[i].Key >= key
  })
}

func (leaf *BpTreeLeaf)Get(key string) (string, bool) {
  i := pageEntryFirstGE(leaf.entries, key)
  if i > 0 {
    //return leaf.entries[i].page.Get(key)
    i--
  }
  return leaf.entries[i].page.Get(key)
}  

func (leaf *BpTreeLeaf)Put(key, value string, fanout int) bool {
  i := pageEntryFirstGE(leaf.entries, key)
  if i != 0 {
    i--
    //only should happen when first is equal to put
    //if greater than should be handled in above tree
  }
  page := leaf.entries[i].page
  inserted := page.Put(key, value)
  
  if inserted && page.Full(fanout) {
    newKey, newPage := page.Split(fanout)
    //TODO: create new pageEntry and add
    //add full and split function for bpTree Leaf, then go up
    //need slicePut helper for []Entries
    //need to add nItem on node types
  
}
*/
