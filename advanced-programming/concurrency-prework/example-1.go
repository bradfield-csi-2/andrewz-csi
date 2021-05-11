package main

import (
	"fmt"
  "sync"
	//"time"
)

func main() {
  var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
    wg.Add(1)
		go func(i int) {
      defer wg.Done()
			fmt.Printf("launched goroutine %d\n", i)
		}(i)
	}
  wg.Wait()
	// Wait for goroutines to finish
	//time.Sleep(time.Second)
}
