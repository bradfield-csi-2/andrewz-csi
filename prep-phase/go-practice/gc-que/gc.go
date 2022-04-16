package main

import (
	"fmt"
	"runtime"
)

// Function to force computation and suppress compiler optimizations
func printSum(a []int) {
	sum := 0
	for i := 0; i < len(a); i++ {
		sum += a[i]
	}
	fmt.Println("\t\tsum calculation (ignore this)", sum)
}

func printHeapUsage(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("HeapAlloc = %v\tlabel: %v\n", m.HeapAlloc, label)
}

func arrayTest(n int) {
	runtime.GC()
	printHeapUsage("begin arrayTest")

	a := make([]int, n)
	for i := 0; i < n; i++ {
		a[i] = i
	}

	runtime.GC()
	printHeapUsage("done allocating array")

	printSum(a)

	// TODO: Try all of these
	//a = nil
	//a = a[500000:]
	//a = a[999999:]
	a = a[len(a):]

	runtime.GC()
	printHeapUsage("done resizing array")

	printSum(a)
}

func main() {
	arrayTest(1000000)

	runtime.GC()
	printHeapUsage("done with arrayTest")
}
