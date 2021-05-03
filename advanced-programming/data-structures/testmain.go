package main

import (
	"fmt"
	"unsafe"
  "reflect"
)


func main() {

	m := make(map[int]int, 10)

  m[30] = 35
  m[62] = 21

  //sk, sv :=  sumMapKV(m)

  //fmt.Printf("keysum = %d | valsum = %d \n", sk, sv)
  ty := reflect.TypeOf(m)

  fmt.Println(ty.String())
  fmt.Println(unsafe.Sizeof(m))


}

