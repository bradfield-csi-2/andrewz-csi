package table

import (
	"compress/zlib"
	"encoding/binary"
	"io"
	"sort"
	"unsafe"
	//"compress/gzip"
)

//#Index

type indexBlock struct {
	indexBlockHeader
	structs []indexBlockStruct
	bytes   []byte
}

type indexBlockHeader struct {
	NBlocks    uint32
	ByteOffset uint32
	ByteLen    uint32
	//Size       uint32
}

type blockIndex struct {
	pages []indexPage
}

type indexPage struct {
	//startKey   string
	endKey     string
	blockStart uint32
}

type indexBlockStruct struct {
	//StartKeyOffset uint32
	//StartKeyLen    uint16
	EndKeyOffset uint32
	EndKeyLen    uint16
	BlockOffset  uint32 //max 4GB file -- need 64B in future?//or if aligned can chunk at 4KB blocks
}

type indexPageKeys struct {
	start string
	end   string
}

func loadIndex(rs io.ReadSeeker, indexOffset int64) (blockIndex, error) {
	index := blockIndex{}
	rs.Seek(indexOffset, io.SeekStart)

	compressReader, err := zlib.NewReader(rs)
	var indexHeader indexBlockHeader
	err = binary.Read(compressReader, binary.LittleEndian, &indexHeader)
	if err != nil {
		return index, err
	}

	structs := make([]indexBlockStruct, indexHeader.NBlocks)
	bytes := make([]byte, indexHeader.ByteLen)
	err = binary.Read(compressReader, binary.LittleEndian, structs)
	if err != nil {
		return index, err
	}

	err = binary.Read(compressReader, binary.LittleEndian, bytes)
	if err != nil {
		return index, err
	}

	strData := string(bytes)
	pages := make([]indexPage, indexHeader.NBlocks)

	for i, idxStruct := range structs {
		//startKeyOffset := int(idxStruct.StartKeyOffset)
		endKeyOffset := int(idxStruct.EndKeyOffset)
		//end := startKeyOffset + int(idxStruct.StartKeyLen)
		end := endKeyOffset + int(idxStruct.EndKeyLen)
		pages[i] = indexPage{
			strData[endKeyOffset:end],
			idxStruct.BlockOffset,
		}
	}

	index.pages = pages
	return index, nil

}

func (bIdx *blockIndex) GetFirstGE(key string) int {
	//n := len(bIdx.pages)
	return sort.Search(len(bIdx.pages), func(i int) bool {
		return bIdx.pages[i].endKey >= key
	})
}

func newIndexBlock(structs []indexBlockStruct, bytes []byte, totKeyLen int) *indexBlock {
	hdr := indexBlockHeader{}
	byteOffset := len(structs)*int(unsafe.Sizeof(structs[0])) + totKeyLen
	hdr.NBlocks = uint32(len(structs))
	hdr.ByteOffset = uint32(byteOffset)
	hdr.ByteLen = uint32(len(bytes))
	return &indexBlock{hdr, structs, bytes}
}

func (idx *indexBlock) Write(w io.Writer) error {
	compressWriter, err := zlib.NewWriterLevel(w, zlib.BestCompression)
	//compressWriter, err := gzip.NewWriterLevel(w,gzip.BestCompression)
	defer compressWriter.Close()

	if err != nil {
		return err
	}

	err = binary.Write(compressWriter, binary.LittleEndian, idx.indexBlockHeader)
	if err != nil {
		return err
	}

	err = binary.Write(compressWriter, binary.LittleEndian, idx.structs)
	if err != nil {
		return err
	}
	err = binary.Write(compressWriter, binary.LittleEndian, idx.bytes)
	if err != nil {
		return err
	}

	err = compressWriter.Flush()
	if err != nil {
		return err
	}

	return nil

}

/*
func getStartKey(items []Item) string {
	return items[0].Key
}
*/

func getEndKey(items []Item) string {
	ilen := len(items)
	return items[ilen-1].Key
}
