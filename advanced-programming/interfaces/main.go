package main

import (
  "fmt"
  "math"
  "runtime"
  "unsafe"
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

// name is an encoded type name with optional extra data.
// See reflect/type.go for details.
type name struct {
	bytes *byte
}

type imethod struct {
	name nameOff
	ityp typeOff
}

type interfacetype struct {
	typ     _type
	pkgpath name
	mhdr    []imethod
}

type itab struct {
	inter *interfacetype
	_type *_type
	hash  uint32 // copy of _type.hash. Used for type switches.
	_     [4]byte
	fun   [1]uintptr // variable sized. fun[0]==0 means _type does not implement inter.
}

type iface struct {
	tab  *itab
	data unsafe.Pointer
}

type geometry interface {
    area() float64
    perim() float64
}

type rect struct {
    width, height float64
}
type circle struct {
    radius float64
}

func main() {
  fmt.Println("Hello")

  var fir interface{} = int(20)
  var sec interface{} = uint64(100)
  m := 30
  var thi interface{} = &m
  one, ok := getIfaceVal(fir)
  if ok {
    fmt.Printf("one = %d \n", one)
  }
  two, ok := getIfaceVal(sec)
  if ok {
    fmt.Printf("two = %d \n", two)
  }
  thr, ok := getIfaceVal(thi)
  if ok {
    fmt.Printf("thr = %d \n", thr)
  }
  r := rect{width: 3, height: 4}
  c := circle{radius: 5}

  printImethods(c)
  printImethods(r)
}


func getIfaceVal(inter interface{}) (val int, ok bool) {
  var dummy interface{} = int(1)
  //isInt := uintptr(*(*unsafe.Pointer)(unsafe.Pointer(&dummy))) == uintptr(*(*unsafe.Pointer)(unsafe.Pointer(&inter)))
  isInt := (*(*iface)(unsafe.Pointer(&dummy))).tab == (*(*iface)(unsafe.Pointer(&inter))).tab
  if isInt {
    //return **(**int)(unsafe.Pointer(uintptr(unsafe.Pointer(&inter)) + unsafe.Offsetof((*(*iface)(unsafe.Pointer(&inter))).data))), true
    ifaceStruct := *(*iface)(unsafe.Pointer(&inter))
    return *(*int)(unsafe.Pointer(ifaceStruct.data)), true
  }
  return -1, false
}

func (r rect) area() float64 {
    return r.width * r.height
}
func (r rect) perim() float64 {
    return 2*r.width + 2*r.height
}

func (c circle) area() float64 {
    return math.Pi * c.radius * c.radius
}
func (c circle) perim() float64 {
    return 2 * math.Pi * c.radius
}

func measure(g geometry) {
    fmt.Println(g)
    fmt.Println(g.area())
    fmt.Println(g.perim())
}

func printImethods(g geometry) {
  ifs := *(*iface)(unsafe.Pointer(&g))
  //ift := *ifs.tab
  //iftyp := *ift.inter//(*interfacetype)(unsafe.Pointer(ift.inter))
  //imethods := ifs.tab.inter.mhdr//iftyp.mhdr

  //imlen := len(imethods)
  //fmt.Printf("imlen = %d\n", imlen)
  
  mlen := len(ifs.tab.inter.mhdr)
  base := &ifs.tab.fun
  for i := 0; i < mlen; i++ {
    funAddr := (*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(base)) + uintptr(i) * unsafe.Sizeof(uintptr(0))))
    fun := runtime.FuncForPC(*funAddr)
    fmt.Printf("Name: %s\n", fun.Name())
    fmt.Printf("Entry:  %x\n", fun.Entry())
    file, line := fun.FileLine(*funAddr)
    fmt.Printf("FileLine:  file = %s | line = %d\n\n",file, line) 
  }

}
