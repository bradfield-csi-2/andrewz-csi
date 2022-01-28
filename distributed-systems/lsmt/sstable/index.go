package sstable

import (
	"compress/lzw"
	"encoding/binary"
	"io"
	"sort"
	"unsafe"
)

//#Index
type IndexRecord struct {
	StartKey   string
	BlockStart uint32
	NItems     uint32
}

type Index struct {
	Records []IndexRecord
}

func loadIndex(rs io.ReadSeeker, indexOffset int64) (*Index, error) {
	//u32buf := make([]byte, 4)
	index := Index{}
	_, err := rs.Seek(indexOffset, io.SeekStart)
	if err != nil {
		return &index, err
	}

	lzwreader := lzw.NewReader(rs, lzw.LSB, 8)
	defer lzwreader.Close()
	var numrecs uint32
	err = binary.Read(lzwreader, binary.LittleEndian, &numrecs)
	if err != nil {
		return &index, err
	}

	index.Records = make([]IndexRecord, numrecs)

	buf := make([]byte, MAX_KEYLEN)

	for i := uint32(0); i < numrecs; i++ {
		rec, err := readRecord(lzwreader, buf)
		if err != nil {
			return &index, err
		}
		index.Records[i] = rec
	}

	return &index, nil

}

func readRecord(r io.Reader, buf []byte) (IndexRecord, error) {
	var keyLen uint32
	var rec IndexRecord
	err := binary.Read(r, binary.LittleEndian, &keyLen)
	if err != nil {
		return rec, err
	}
	if keyLen > MAX_KEYLEN {
		panic("crazy long key read in record")
	}

	err = binary.Read(r, binary.LittleEndian, buf[:keyLen])
	if err != nil {
		return rec, err
	}
	rec.StartKey = string(buf[:keyLen])

	err = binary.Read(r, binary.LittleEndian, &rec.BlockStart)
	if err != nil {
		return rec, err
	}

	err = binary.Read(r, binary.LittleEndian, &rec.NItems)
	return rec, err

}

func writeIndex(w io.Writer, recs []IndexRecord) error {
	n := uint32(len(recs))
	err := binary.Write(w, binary.LittleEndian, n)
	if err != nil {
		return err
	}
	for _, rec := range recs {
		err = writeRecord(w, rec)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeRecord(w io.Writer, rec IndexRecord) error {
	keyLen := uint32(len(rec.StartKey))
	err := binary.Write(w, binary.LittleEndian, keyLen)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, *(*[]byte)(unsafe.Pointer(&rec.StartKey)))
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, rec.BlockStart)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, rec.NItems)

	return nil
}

func (idx *Index) getLastLE(key string) int {
	n := len(idx.Records)
	i := sort.Search(n, func(i int) bool {
		return idx.Records[i].StartKey >= key
	})
	if i < n && idx.Records[i].StartKey == key {
		return i
	}
	return i - 1
}
