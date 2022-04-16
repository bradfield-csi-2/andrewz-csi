package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
)

const (
	EQUALS = iota
	GT
	LT
)

const (
	ASC = iota
	DESC
)

/*
func main() {
	t := getTest()
	t.Init()
	r, err := t.Next()
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	for i := 0; i < 4; i++ {
		r, err = t.Next()
		fmt.Println(r)
	}

	r, err = t.Next()
	if err != io.EOF {
		panic("wrong")
	}
	t.Close()
}
*/
/*
func getTest() QueryNode {
	scan := makeFileScanNode("movies") //TODO: make real file scan now
	//sel := selectNode{predicate{0, EQUALS, "200"}, &scan}
	//proj := projectNode{[]int{1}, &sel}
	//sort := sortNode{dir: DESC, fields: []int{0}, input: &scan}
	proj := makeProjNode([]int{0, 1}, 3, &scan) //projectNode{[]int{0, 1}, &sort}
	lim := makeLimNode(5, proj)
	//sel.input = &scan
	//proj.input = &sel
	return lim //proj
}
*/

type QueryNode interface {
	Init()
	HasNext() bool
	Execute() record
	Close()
}

type sortNode struct {
	dir    int
	fields []int
	buf    []record
	cursor int
	end    int
	input  QueryNode
	next   record
}

func (n *sortNode) Init() {
	n.input.Init()

	for hasnext := n.input.HasNext(); hasnext; hasnext = n.input.HasNext() {

		n.buf = append(n.buf, n.input.Execute())

	}
	sort.Slice(n.buf, func(i, j int) bool {
		for _, f := range n.fields {
			if n.buf[i][f] == n.buf[j][f] {
				continue
			} else if n.dir == ASC {
				return n.buf[i][f] < n.buf[j][f]
			} else {
				return n.buf[i][f] > n.buf[j][f]
			}
		}
		return true
	})
	n.cursor = 0
	n.end = len(n.buf)
}

func (n *sortNode) Execute() record {
	return n.next
}

func (n *sortNode) HasNext() bool {
	if n.cursor >= n.end {
		return false
	}
	n.next = n.buf[n.cursor]
	n.cursor++
	return true
}

func (n *sortNode) Close() {
	n.input.Close()
}

type limitNode struct {
	limit int
	count int
	input QueryNode
	next  record
}

func makeLimNode(limit int, input QueryNode) *limitNode {
	return &limitNode{limit: limit, input: input}
}

func (n *limitNode) Init() {
	n.input.Init()
	n.count = 0
}

func (n *limitNode) Execute() record {
	return n.next
}

func (n *limitNode) HasNext() bool {
	if n.count >= n.limit || !n.input.HasNext() {
		return false
	}
	n.next = n.input.Execute()
	n.count++
	return true
}

func (n *limitNode) Close() {
	n.input.Close()
}

type projectNode struct {
	fields []int
	ncols  int
	input  QueryNode
	next   record
}

func makeProjNode(fields []int, ncols int, input QueryNode) *projectNode {
	return &projectNode{fields: fields, ncols: ncols, input: input}
}

func (n *projectNode) Init() {
	n.input.Init()
}

func (n *projectNode) Execute() record {
	return n.next
}

func (n *projectNode) HasNext() bool {
	if !n.input.HasNext() {
		return false
	}
	rec := n.input.Execute()
	prec := make([]byte, len(n.fields)<<1) //size of pointer is 2
	c := 0
	//end := len(prec)
	//need schema
	for _, f := range n.fields {
		//copy(prec[c << 1:(c << 1) + 2], rec[f << 1: (f <<1) + 2])
		fieldVal := getFieldVal(rec, n.ncols, f)

		binary.LittleEndian.PutUint16(prec[c<<1:], uint16(len(prec)))
		prec = append(prec, fieldVal...)
		c++
	}
	n.next = prec
	return true
}

func (n *projectNode) Close() {
	n.input.Close()
}

type predicate struct {
	field int
	op    int
	param []byte
}

type selectNode struct {
	pred  predicate
	ncols int
	input QueryNode
	next  record
}

func makeSelNode(pred predicate, ncols int, input QueryNode) *selectNode {
	return &selectNode{pred: pred, ncols: ncols, input: input}
}

func (n *selectNode) Init() {
	n.input.Init()
}

func (n *selectNode) Execute() record {
	return n.next
}

func (n *selectNode) HasNext() bool {

	for {
		if !n.input.HasNext() {
			return false
		}
		rec := n.input.Execute()
		//val := rec[n.pred.field]
		val := getFieldVal(rec, n.ncols, n.pred.field)
		switch n.pred.op {
		case EQUALS:
			if ByteSliceEqual(val, n.pred.param) {
				n.next = rec
				return true
			}
			continue
		default:
			fmt.Println(n.pred.op)
			panic("Predicat OP not handled")
		}
	}

}

func (n *selectNode) Close() {
	n.input.Close()
}

func makeFileScanNode(table string) *fileScanNode {
	return &fileScanNode{table: table}
}

type fileScanNode struct {
	//records    []record
	table    string
	slot     int
	blockId  int64
	blockEnd int64
	next     record
}

func (n *fileScanNode) Init() {
	//n.records = make([]record, 0) //initTable()
	//get table location
	filepath := TblMgr.GetTableFilepath(n.table)
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	info, err := f.Stat()
	if err != nil {
		panic(err)
	}

	n.slot = 0
	n.blockId = 0
	n.blockEnd = info.Size() / BUFPAGESIZE
}

func (n *fileScanNode) Execute() record {
	return n.next
}

func (n *fileScanNode) HasNext() bool {
	if n.blockId >= n.blockEnd {
		return false
	}
	bufpg := BufMgr.GetBufPage(n.table, n.blockId)
	for bufpg.HasRecord(n.slot) {
		n.next = bufpg.GetRecord(n.slot)
		n.slot++
		if RecIsValid(n.next) {
			bufpg.Release()
			return true
		}

	}
	n.blockId++
	n.slot = 0
	bufpg.Release()
	if n.blockId >= n.blockEnd {
		return false
	}
	bufpg = BufMgr.GetBufPage(n.table, n.blockId)
	for bufpg.HasRecord(n.slot) {
		n.next = bufpg.GetRecord(n.slot)
		n.slot++
		if RecIsValid(n.next) {
			bufpg.Release()
			return true
		}
	}
	bufpg.Release()
	return false

}

func (n *fileScanNode) Close() {
	//empty for now
	//in future can close file
}

type movie struct {
	movieId int
	title   string
	genres  string
}

type record []byte

func RecIsValid(r record) bool {
	return len(r) > 0 && (r[0] == 0)
}

func (a *record) Equal(b record) bool {
	if len(*a) != len(b) {
		return false
	}
	n := len(*a)
	for i := 0; i < n; i++ {
		if (*a)[i] != b[i] {
			return false
		}
	}
	return true
}

func getFieldVal(rec record, ncols, fieldCol int) []byte {
	fstart := binary.LittleEndian.Uint16(rec[fieldCol<<1:])
	var fend uint16
	if fieldCol == ncols-1 {
		fend = uint16(len(rec))
	} else {
		fend = binary.LittleEndian.Uint16(rec[(fieldCol+1)<<1:])
	}
	return rec[fstart:fend]
}

func ByteSliceEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	n := len(a)
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
