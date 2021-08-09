package bitmap

//import "fmt"
import "sort"

const WordSize = 64
const WSzBits = 16
const LSzBits = 4
const LowBitMask16 = 0x0ffff
const LowBitMask4 = 0x0f
const MaxArraySize = 4096
const BitmapSize = (1 << 12)

const (
  Array = iota
  Bitmap
)

type roaringBitmap struct {
	data []roarBucket
}

type roarBucket struct {
  typ uint8
  //n   uint16
  bits []uint16
}


type nums []uint16

func (n nums) Len() int           { return len(n) }
func (n nums) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n nums) Less(i, j int) bool { return n[i] < n[j] }

func newRoaringBitmap() *roaringBitmap {
	return &roaringBitmap{}
}

func (b *roaringBitmap) Get(x uint32) bool {
  idx := x >> WSzBits
  if int(idx) >= len(b.data) {
    return false
  }
  bucket := b.data[idx]

  low16Bits := uint16(x & LowBitMask16)
  if bucket.typ == Array {
    for _, ele := range bucket.bits {
      if ele ==  low16Bits {
        return true
      }
    }
    return false
  }

  idx = uint32(low16Bits >> LSzBits)
  if int(idx) >= len(bucket.bits) {
    return false
  }
  shiftAmt := uint16(low16Bits & LowBitMask4)
 
  entryMask := uint16(1) << shiftAmt
	return (bucket.bits[idx] & entryMask) != 0
}



func (b *roaringBitmap) Set(x uint32) {
  idx := x >> WSzBits
  if int(idx) >= len(b.data) {
    dataExt := make([]roarBucket,int(idx) - len(b.data) + 1)
    b.data = append(b.data, dataExt...)
  }
  bucket := &(b.data[idx])
  low16Bits := uint16(x & LowBitMask16)
  if bucket.typ == Array {
    n := len(bucket.bits)//int(bucket.n)
    i := sort.Search(n, func(i int) bool { return bucket.bits[i] >= low16Bits })

    if i < n && bucket.bits[i] == low16Bits {
      return
    }

    bucket.bits = append(bucket.bits, low16Bits)

    if i < n {
      for j := n; j > 0; j-- {
        if j == i {
          break
        }
        bucket.bits[j] = bucket.bits[j-1]
      }
    }
    bucket.bits[i] = low16Bits
    //bucket.n++
 
    if len(bucket.bits) > MaxArraySize {
      bucket.convertArray2Bitmap()

    } 
    return
  }
 
  idx = uint32(low16Bits >> LSzBits)
  shiftAmt := uint16(low16Bits & LowBitMask4)
 
  entryMask := uint16(1) << shiftAmt
	bucket.bits[idx]  = (bucket.bits[idx] | entryMask)

}


func (bucket *roarBucket)convertArray2Bitmap() {
  bucket.typ = Bitmap
  replace := make([]uint16,BitmapSize) 

  var shiftAmt, entryMask uint16
  var k uint32
  for _, ele := range bucket.bits {
    k = uint32(ele >> LSzBits)
    shiftAmt = uint16(ele & LowBitMask4)
    entryMask = uint16(1) << shiftAmt
    replace[k]  = (replace[k] | entryMask)
  }
  bucket.bits = replace
  bucket.typ = Bitmap
	 
}

/**/
func (b *roaringBitmap) Union(other *roaringBitmap) *roaringBitmap {
	var data, less, more []roarBucket
  if len(b.data) > len(other.data) {
    less = other.data
    more = b.data
  } else {
    less = b.data
    more = other.data
  }
  data = make([]roarBucket, len(more))
  i := 0
  lLen := len(less)
  for i < lLen {
    a, b := &(less[i]), &(more[i])
    data[i] = unionBuckets(a,b)
  }
  mLen := len(more)
  for i < mLen {
    data[i] = more[i]
    data[i].bits = make([]uint16, BitmapSize)
    copy(data[i].bits, more[i].bits)
    i++
  }
    
	return &roaringBitmap{
		data: data,
	}
}


func unionBuckets(a, b *roarBucket) roarBucket {
  //caseNum := 0
  var bucket roarBucket
  typSum := a.typ + b.typ
  switch typSum {
  case 0:
    //TODO: both are arrays
    bucket = mergeArrays(a.bits, b.bits)
  case 1:
    //TODO: just one is array
    bucket = unionArrayIntoBitmap(a, b)
  case 2:
    //TODO: both are bitmaps
    bucket = *a
    bucket.bits = make([]uint16, BitmapSize)
    for i:=0; i < BitmapSize; i++ {
      bucket.bits[i] = a.bits[i] | b.bits[i]
    }
  default:
    panic("case not handled")
  }
  return bucket
}


func unionArrayIntoBitmap(a, b *roarBucket) roarBucket {
  var arr, bmap []uint16
  var bucket roarBucket
  if a.typ == Array {
    bucket = *b
    arr, bmap = a.bits, b.bits
  } else {
    bucket = *a
    arr, bmap = b.bits, a.bits
  }

  bucket.bits = make([]uint16, BitmapSize)
  copy(bucket.bits, bmap)

  for _, num := range arr {
    idx := num >> LSzBits
    shiftAmt := uint16(num & LowBitMask4)
 
    entryMask := uint16(1) << shiftAmt
    bucket.bits[idx] |=  entryMask
  }
  return bucket
}

func mergeArrays(a, b []uint16) roarBucket {
  /*
  var less, more uint16
  if len(a.bits) < len(b.bits) {
    less, more = a.bits, b.bits
  } else {
    less, more = b.bits, a.bits
  }
  */

  bits := make([]uint16, 0)
  bucket := roarBucket{typ:Array}
  i, j := 0,0
  for i < len(a) && j < len(b) {
    if a[i] < b[j] {
      bits = append(bits, a[i])
      i++
    } else if a[j] < b[i] {
      bits = append(bits, b[j])
    } else {
      bits = append(bits, a[i])
      i++
      j++
    }
  }

  if i < len(a) {
    bits = append(bits, a[i:]...)
  } else if j < len(b) {
    bits = append(bits,b[j:]...)
  }

  bucket.bits = bits
  if len(bits) > MaxArraySize {
    //TODO: convert to bitmap
    bucket.convertArray2Bitmap()
  }

  return bucket
}
/**/
/*
func (b *uncompressedBitmap) Intersect(other *uncompressedBitmap) *uncompressedBitmap {
	//var data []uint64
	var  less, more []uint64
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
 
	return &uncompressedBitmap{
		data: less,//data,
	}
}

  */
