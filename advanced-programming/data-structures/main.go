package main

import (
	"fmt"
	"unsafe"
  "reflect"
)


const (
	// Maximum number of key/elem pairs a bucket can hold.
	bucketCntBits = 3
	bucketCnt     = 1 << bucketCntBits

	// Maximum average load of a bucket that triggers growth is 6.5.
	// Represent as loadFactorNum/loadFactorDen, to allow integer math.
	loadFactorNum = 13
	loadFactorDen = 2

	// Maximum key or elem size to keep inline (instead of mallocing per element).
	// Must fit in a uint8.
	// Fast versions cannot handle big elems - the cutoff size for
	// fast versions in cmd/compile/internal/gc/walk.go must be at most this elem.
	maxKeySize  = 128
	maxElemSize = 128

	// data offset should be the size of the bmap struct, but needs to be
	// aligned correctly. For amd64p32 this means 64-bit alignment
	// even though pointers are 32 bit.
	dataOffset = unsafe.Offsetof(struct {
		b bmap
		v int64
	}{}.v)

	// Possible tophash values. We reserve a few possibilities for special marks.
	// Each bucket (including its overflow buckets, if any) will have either all or none of its
	// entries in the evacuated* states (except during the evacuate() method, which only happens
	// during map writes and thus no one else can observe the map during that time).
	emptyRest      = 0 // this cell is empty, and there are no more non-empty cells at higher indexes or overflows.
	emptyOne       = 1 // this cell is empty
	evacuatedX     = 2 // key/elem is valid.  Entry has been evacuated to first half of larger table.
	evacuatedY     = 3 // same as above, but evacuated to second half of larger table.
	evacuatedEmpty = 4 // cell is empty, bucket is evacuated.
	minTopHash     = 5 // minimum tophash for a normal filled cell.

	// flags
	iterator     = 1 // there may be an iterator using buckets
	oldIterator  = 2 // there may be an iterator using oldbuckets
	hashWriting  = 4 // a goroutine is writing to the map
	sameSizeGrow = 8 // the current map growth is to a new map of the same size

	// sentinel bucket ID for iterator checks
	//noCheck = 1<<(8*sys.PtrSize) - 1
)

type tflag uint8

type nameOff int32
type typeOff int32

type _type struct {
	size       uintptr
	ptrdata    uintptr // size of memory prefix holding all pointers
	hash       uint32
	tflag      tflag
	align      uint8
	fieldAlign uint8
	kind       uint8
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	equal func(unsafe.Pointer, unsafe.Pointer) bool
	// gcdata stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, gcdata is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	gcdata    *byte
	str       nameOff
	ptrToThis typeOff
}

type maptype struct {
	typ    _type
	key    *_type
	elem   *_type
	bucket *_type // internal type representing a hash bucket
	// function for hashing keys (ptr to key, seed) -> hash
	hasher     func(unsafe.Pointer, uintptr) uintptr
	keysize    uint8  // size of key slot
	elemsize   uint8  // size of elem slot
	bucketsize uint16 // size of bucket
	flags      uint32
}


// isEmpty reports whether the given tophash array entry represents an empty bucket entry.
func isEmpty(x uint8) bool {
	return x <= emptyOne
}

// A header for a Go map.
type hmap struct {
	// Note: the format of the hmap is also encoded in cmd/compile/internal/reflectdata/reflect.go.
	// Make sure this stays in sync with the compiler's definition.
	count     int // # live cells == size of map.  Must be first (used by len() builtin)
	flags     uint8
	B         uint8  // log_2 of # of buckets (can hold up to loadFactor * 2^B items)
	noverflow uint16 // approximate number of overflow buckets; see incrnoverflow for details
	hash0     uint32 // hash seed

	buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0.
	oldbuckets unsafe.Pointer // previous bucket array of half the size, non-nil only when growing
	nevacuate  uintptr        // progress counter for evacuation (buckets less than this have been evacuated)

	extra *mapextra // optional fields
}

// mapextra holds fields that are not present on all maps.
type mapextra struct {
	// If both key and elem do not contain pointers and are inline, then we mark bucket
	// type as containing no pointers. This avoids scanning such maps.
	// However, bmap.overflow is a pointer. In order to keep overflow buckets
	// alive, we store pointers to all overflow buckets in hmap.extra.overflow and hmap.extra.oldoverflow.
	// overflow and oldoverflow are only used if key and elem do not contain pointers.
	// overflow contains overflow buckets for hmap.buckets.
	// oldoverflow contains overflow buckets for hmap.oldbuckets.
	// The indirection allows to store a pointer to the slice in hiter.
	overflow    *[]*bmap
	oldoverflow *[]*bmap

	// nextOverflow holds a pointer to a free overflow bucket.
	nextOverflow *bmap
}

// A bucket for a Go map.
type bmap struct {
	// tophash generally contains the top byte of the hash value
	// for each key in this bucket. If tophash[0] < minTopHash,
	// tophash[0] is a bucket evacuation state instead.
	tophash [bucketCnt]uint8
	// Followed by bucketCnt keys and then bucketCnt elems.
	// NOTE: packing all the keys together and then all the elems together makes the
	// code a bit more complicated than alternating key/elem/key/elem/... but it allows
	// us to eliminate padding which would be needed for, e.g., map[int64]int8.
	// Followed by an overflow pointer.
}

// A hash iteration structure.
// If you modify hiter, also change cmd/compile/internal/reflectdata/reflect.go to indicate
// the layout of this structure.
/*
type hiter struct {
	key         unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/compile/internal/walk/range.go).
	elem        unsafe.Pointer // Must be in second position (see cmd/compile/internal/walk/range.go).
	t           *maptype
	h           *hmap
	buckets     unsafe.Pointer // bucket ptr at hash_iter initialization time
	bptr        *bmap          // current bucket
	overflow    *[]*bmap       // keeps overflow buckets of hmap.buckets alive
	oldoverflow *[]*bmap       // keeps overflow buckets of hmap.oldbuckets alive
	startBucket uintptr        // bucket iteration started at
	offset      uint8          // intra-bucket offset to start from during iteration (should be big enough to hold bucketCnt-1)
	wrapped     bool           // already wrapped around from end of bucket array to beginning
	B           uint8
	i           uint8
	bucket      uintptr
	checkBucket uintptr
}
*/

func main() {
	fmt.Println("Hello")
	var f float64 = 20.5
	b := getFloatBin(f)
	fmt.Printf("%x\n", b)
	fmt.Printf("%064b\n", b)
	
	s, u := "s string", "u string"
	t := s
	s,u,t = s[1:],u[1:],t[1:]

	compS := isSameString(s, t)
	compU := isSameString(s, u)

	fmt.Printf("%v\n", compS)
	fmt.Printf("%v\n", compU)

	bs := []int{1,2,3,4,5,6,7,8,9,10}
	sum := unsafeSum(bs)
	fmt.Printf("%d\n", sum)

	m := make(map[int]int, 10)
	one := checkLen(m)
	m[1] = 2
	two := checkLen(m)

	fmt.Printf("one is %v\n", one)
	fmt.Printf("two is %v\n", two)

  m[30] = 35
  m[62] = 21

  sk, sv :=  sumMapKV(m)

  fmt.Printf("keysum = %d | valsum = %d \n", sk, sv)
  ty := reflect.TypeOf(m)

  fmt.Println(ty.String())
  fmt.Println(unsafe.Sizeof(m))


}

func getFloatBin(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}


func isSameString(s, t string) bool {
	sp := *(**[]byte)(unsafe.Pointer(&s))
	tp := *(**[]byte)(unsafe.Pointer(&t))
	return tp == sp
}

func unsafeSum(s []int) int {
	sum := 0
	//slen := len(s)
	sp := *(**int)(unsafe.Pointer(&s))
	ep := (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(sp)) + unsafe.Sizeof(*sp) * uintptr(len(s))))
	for uintptr(unsafe.Pointer(sp)) < uintptr(unsafe.Pointer(ep)) {
		sum += *sp
		sp = (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(sp)) + unsafe.Sizeof(*sp)))
	}
	return sum
}


func checkLen(m map[int]int) bool {
	mLen := len(m)
	uLen := **(**int)(unsafe.Pointer(&m))
	fmt.Printf("mLen = %d\n", mLen)
	fmt.Printf("uLen = %d\n", uLen)
  fmt.Printf("%v\n", unsafe.Sizeof(1))
	return mLen == uLen
}


func sumMapKV(m map[int]int) (sk, sv int) {
  //bucketsize is 144 data offset = 8, overflow offset = 136
  //
  //
  //get map hmap
  //get buckets
  //get count
  //for earch bucket which is not empty
  //get the key and value
  //add to running sum
  //search all overflow buckets
  // no overflow if overflow pointer is nil??
  sk, sv = 0,0
  hm := **(**hmap)(unsafe.Pointer(&m))

  b := hm.buckets
  bucket := 0
  sCnt := 0
  nbuckets := 1 << hm.B
  fmt.Println("hm.B")
  fmt.Println(hm.B)
next:
  //
  if sCnt == hm.count {
    return
  }
  if bucket >= nbuckets {
    fmt.Println(bucket)
    fmt.Println(nbuckets)
    panic("out of bucket bounds")
  }


  bm := *(*bmap)(b)

  karr := *(*[8]int)(unsafe.Pointer(uintptr(b) + unsafe.Sizeof(bm)))
  earr := (*[8]int)(unsafe.Pointer(uintptr(b) + unsafe.Sizeof(bm) + unsafe.Sizeof(karr)))
  for i := 0; i < bucketCnt; i++ {
		//offi := (i + it.offset) & (bucketCnt - 1)
		if isEmpty(bm.tophash[i])  {
      continue
    }

    sk += karr[i]
    sv += earr[i]
    sCnt++
  }

  b = unsafe.Pointer(*(**bmap)(unsafe.Pointer(uintptr(b) + uintptr(144-8))))

  fmt.Println("check nil overflow")
  if (*bmap)(b) == nil {
    bucket++
    b = unsafe.Pointer(uintptr(hm.buckets) + uintptr(bucket) * uintptr(144))
    fmt.Println("next bucket")
  }

  goto next

}


