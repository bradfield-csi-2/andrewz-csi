package main

/*
#include <stdlib.h>

int square(int x) {
  return x * x;
}

*/
import "C"
import  "fmt"

func Random() int {
    return int(C.random())
}

func Seed(i int) {
    C.srandom(C.uint(i))
}


func main() {
  fmt.Println("hellow")
  s := int(C.square(C.int(3)))
  fmt.Println(s)
}
