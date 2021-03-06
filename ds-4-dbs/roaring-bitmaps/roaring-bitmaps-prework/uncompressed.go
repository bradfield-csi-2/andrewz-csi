package bitmap

const wordSize = 64
const wSzBits = 6
const lowBitMask = 0x03f

type uncompressedBitmap struct {
	data []uint64
}

func newUncompressedBitmap() *uncompressedBitmap {
	return &uncompressedBitmap{make([]uint64, 1024)}
}

func (b *uncompressedBitmap) Get(x uint32) bool {
	idx := x >> wSzBits
	if int(idx) >= len(b.data) {
		return false
	}
	shiftAmt := uint64(x & lowBitMask)
	entryMask := uint64(1) << shiftAmt
	return (b.data[idx] & entryMask) != 0
}

func (b *uncompressedBitmap) Set(x uint32) {
	idx := x >> wSzBits
	shiftAmt := uint64(x & lowBitMask)
	entryMask := uint64(1) << shiftAmt
	if int(idx) >= len(b.data) {

		if int(idx) >= cap(b.data) {
			extCap := cap(b.data) << 1

			for extCap <= int(idx) {
				extCap <<= 1
			}

			ext := make([]uint64, int(idx)+1, extCap)
			copy(ext, b.data)
			b.data = ext
		} else {
			b.data = b.data[:idx+1]
		}

	}
	b.data[idx] = b.data[idx] | entryMask
}

func (b *uncompressedBitmap) Union(other *uncompressedBitmap) *uncompressedBitmap {
	var data, less, more []uint64
	if len(b.data) > len(other.data) {
		less = other.data
		more = b.data
	} else {
		less = b.data
		more = other.data
	}
	data = make([]uint64, len(more))
	i := 0
	lLen := len(less)
	for i < lLen {
		data[i] = less[i] | more[i]
		i++
	}
	mLen := len(more)
	for i < mLen {
		data[i] = more[i]
		i++
	}

	return &uncompressedBitmap{
		data: data,
	}
}

func (b *uncompressedBitmap) Intersect(other *uncompressedBitmap) *uncompressedBitmap {
	//var data []uint64
	var less, more []uint64
	if len(b.data) > len(other.data) {
		less = other.data
		more = b.data
	} else {
		less = b.data
		more = other.data
	}
	//data = make([]uint64, len(less))
	i := 0
	lLen := len(less)
	for i < lLen {
		less[i] = less[i] & more[i]
		i++
	}
	/*
	  mLen := len(more)
	  for i < mLen {
	    data[i] = more[i]
	  }
	*/

	return &uncompressedBitmap{
		data: less, //data,
	}
}
