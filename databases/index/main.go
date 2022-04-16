package main

import (
	"encoding/csv"
	"fmt"
	"os"
	//"sync/atomic"
)

//var fileTokens chan struct{} = make(chan struct{}, 20)

func main() {
	fmt.Println("hi")
	f, err := os.OpenFile("/Users/andrewzheng/Downloads/ml-20m/movies.csv", os.O_RDONLY, 0)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	csr := csv.NewReader(f)

	rec, _ := csr.Read()

	fmt.Println(rec)

	for i := 0; i < 13; i++ {
		rec, _ := csr.Read()

		fmt.Println(rec)
	}

}

type memschema struct {
	nameColMap map[string]int
	cols       []colType
}

type colType struct {
	typId int
	len   int
}

type block struct {
	table   string
	blockId int
	schema  int
}

//for now just support int
//and string

func protoTypeConv() {
	//one: check type
	//two read binary
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
