package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"testing"
)

/*
type bufPage struct {
	pinCnt   int
	filename string
	blockId  int
	used     uint16
	upper    uint16
	lower    uint16
	rwmutex  sync.RWMutex
	data     []byte
}
*/
/*
func TestReadAndWrite(t *testing.T) {
	b := &bufpage{}
	b.filepath = "./testdb/test"
	b.blockId = 0
	b.data = make([]byte, 4096)
	b.upper = 6
	b.lower = 4096
	b.used = 6
	rec := make([]byte, 7)
	copy(rec, "ulysses")
	b.WriteRecord(rec)
	b.Flush()
	new := b.GetRecord(0)
	if len(new) != len(rec) {
		t.Fatalf("Record Lengths don't match")
	}

	for i := 0; i < len(new); i++ {
		if rec[i] != new[i] {
			t.Fatalf("Record byte doesn't match")
		}
	}

}
*/
func TestReadAllIntoMem(t *testing.T) {
	//set up
	//create test folder and file for the table
	//add records and byte
	tbldir, err := os.MkdirTemp("./testdb/tables", "test")
	if err != nil {
		panic(err)
	}
	_, err = os.Create(tbldir + "/data")

	//records - int, string, string
	movies := getMovies()
	TblMgr = initTblMgr("testdb", ".")

	BufMgr = initBufMgr(20)

	BufMgr.run()
	defer BufMgr.stop()

	tblname := path.Base(tbldir)
	bufpg := BufMgr.GetBufPage(tblname, 0)
	//bufpg.Init(tblname, 0, tm) //should be in manager

	/*
		used := uint16(6)
		upper := uint16(6)
		lower := uint16(BUFPAGESIZE)
		page := make([]byte, BUFPAGESIZE)
	*/
	//fmt.Println(bufpg)
	for _, m := range movies {
		r := createMovieBytesRow(m)
		bufpg.WriteRecord(r)
		/*
			rlen := uint16(len(r))
			lower -= rlen
			copy(page[lower:], r)
			binary.LittleEndian.PutUint16(page[upper:], lower)
			binary.LittleEndian.PutUint16(page[upper+2:], rlen)
			upper += 4
			used += 4 + rlen
		*/
	}
	bufpg.Flush()
	bufpg.Release()
	/*
		binary.LittleEndian.PutUint16(page, used)
		binary.LittleEndian.PutUint16(page[2:], upper)
		binary.LittleEndian.PutUint16(page[4:], lower)

		binary.Write(tbfile, binary.LittleEndian, page)
	*/
	bufpg = BufMgr.GetBufPage(tblname, 0)
	for i, m := range movies {
		r1 := createMovieBytesRow(m)
		//if r == offset
		r2 := bufpg.GetRecord(i)
		if len(r1) != len(r1) {
			t.Fatalf("Record Lengths don't match")
		}
		for j := 0; j < len(r1); j++ {
			if r1[j] != r2[j] {
				t.Fatalf("Record byte doesn't match")
			}
		}

	}
	bufpg.Release()
	os.RemoveAll(tbldir)
	//tear down
}

func TestCopyMoviesCSVToFile(t *testing.T) {
	//set up
	//create test folder and file for the table
	//add records and byte
	fmt.Println("Starting copy csv test")

	tbldir, err := os.MkdirTemp("./testdb/tables", "movies")
	defer os.RemoveAll(tbldir)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created table folder: %s \n", tbldir)
	check, err := os.Create(tbldir + "/data")
	check.Close()

	fmt.Println("Created file")
	//records - int, string, string
	TblMgr = initTblMgr("testdb", ".")
	BufMgr = initBufMgr(20)

	fmt.Println("Initialized table and buffer manager")

	BufMgr.run()
	fmt.Println("Running buffer manager")
	defer BufMgr.stop()

	tblname := path.Base(tbldir)

	f, err := os.OpenFile("/Users/andrewzheng/Downloads/ml-20m/movies.csv", os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("err on opening file")
		os.Exit(0)
	}
	csr := csv.NewReader(f)

	fmt.Println("Reading CSV File")

	rec, err := csr.Read()

	blockid := int64(0)
	fmt.Println("Inserting csv records into db")
	for err != io.EOF {

		bufpg := BufMgr.GetBufPage(tblname, blockid)
		//bufpg.Init(tblname, tbldir+"/data", blockid)

		for slot := uint16(0); bufpg.CanWriteRecord(slot); slot++ {
			rec, err = csr.Read()
			if err == io.EOF {
				break
			}
			mrec := convertStringRecToMovie(rec)
			brec := createMovieBytesRow(mrec)
			bufpg.WriteRecord(brec)

		}
		bufpg.Flush()
		bufpg.Release()
		blockid++

	}

	f.Close()

	fmt.Println("Done inserting records")
	scan := makeFileScanNode(tblname) //TODO: make real file scan now
	fmt.Println("Initializing scan")
	scan.Init()

	f, err = os.OpenFile("/Users/andrewzheng/Downloads/ml-20m/movies.csv", os.O_RDONLY, 0)
	defer f.Close()
	if err != nil {
		t.Fatalf("err on opening file")
		os.Exit(0)
	}
	csr = csv.NewReader(f)

	rec, err = csr.Read()
	row := 1
	fmt.Println("Checking matching records")
	DEBUG = true

	for hasnext := scan.HasNext(); hasnext; hasnext = scan.HasNext() {
		srec := scan.Execute()
		rec, err = csr.Read()
		if err == io.EOF {
			t.Fatalf("csv records ran out before file scan")
		}
		mrec := convertStringRecToMovie(rec)
		brec := createMovieBytesRow(mrec)
		if !ByteSliceEqual([]byte(srec), brec) {
			fmt.Println(srec)
			fmt.Println(brec)
			t.Fatalf("rows  %d are not equal", row)
			os.Exit(0)
		}
		row++
	}
	fmt.Println("Done matching records")

	fmt.Println("tearing down temp folders")
	//tear down
}

type movie_test struct {
	id    uint32
	title string
	genre string
}

func convertStringRecToMovie(record []string) movie_test {
	if len(record) != 3 {
		panic("did not pass three wide rec")
	}
	id, err := strconv.Atoi(record[0])
	if err != nil {
		panic("error on conv movie id string to int")
	}
	return movie_test{uint32(id), record[1], record[2]}
}

func getMovies() []movie_test {
	movies := []movie_test{
		{1, "Toy Story (1995)", "Adventure|Animation|Children|Comedy|Fantasy    "},
		{2, "Jumanji (1995)", "Adventure|Children|Fantasy"},
		{3, "Grumpier Old Men (1995)", "Comedy|Romance"},
		{4, "Waiting to Exhale (1995)", "Comedy|Drama|Romance"},
		{5, "Father of the Bride Part II (1995)", "Comedy"},
	}
	return movies

}

func createMovieBytesRow(m movie_test) record {
	//l := 4 + len(m.title) + len(m.genre) + 2 + 2 + 2
	buf := bytes.NewBuffer(make([]byte, 0))

	err := binary.Write(buf, binary.LittleEndian, uint8(0))

	var ptr uint16 = 7

	//binary.LittleEndian.PutUint16(rec[:], 6)
	err = binary.Write(buf, binary.LittleEndian, ptr)
	if err != nil {
		panic(err)
	}

	ptr += uint16(binary.Size(m.id))
	err = binary.Write(buf, binary.LittleEndian, ptr)
	if err != nil {
		panic(err)
	}
	ptr += uint16(len(m.title))
	err = binary.Write(buf, binary.LittleEndian, ptr)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, binary.LittleEndian, m.id)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, binary.LittleEndian, []byte(m.title))
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, binary.LittleEndian, []byte(m.genre))
	if err != nil {
		panic(err)
	}
	return buf.Bytes()

}
