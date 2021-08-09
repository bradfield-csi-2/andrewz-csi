package bitmap

const (
	Literal = iota
	Fill0
	Fill1
)

const (
	LL = (Literal << 2) | Literal
	LZ = (Literal << 2) | Fill0
	LO = (Literal << 2) | Fill1
	ZL = (Fill0 << 2) | Literal
	ZZ = (Fill0 << 2) | Fill0
	ZO = (Fill0 << 2) | Fill1
	OL = (Fill1 << 2) | Literal
	OZ = (Fill1 << 2) | Fill0
	OO = (Fill1 << 2) | Fill1
)

func compress(b *uncompressedBitmap) []uint64 {
	return wah32(b.data) //wah64(b.data)
}

func decompress(compressed []uint64) *uncompressedBitmap {

	var data []uint64
	data = deWah32(compressed) //deWah64(compressed)

	return &uncompressedBitmap{
		data: data,
	}
}


