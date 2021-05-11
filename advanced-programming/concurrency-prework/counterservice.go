package main


import (
  "fmt"
  "sync/atomic"
  "sync"
)


type counterService interface {
    // Returns values in ascending order; it should be safe to call
    // getNext() concurrently without any additional synchronization.
    getNext() uint64
}

type nosync int

type atom int

type mutx int

type pcntr struct {
  reqs chan uint64
  resps chan uint64
}

var counter uint64

var cntMutex = sync.Mutex{}

//var reqs chan uint64

//var resps chan uint64

func main() {
  counter = 0
  fmt.Println("Start")

  //reqs = make(chan uint64)

  //resps = make(chan uint64)

  pc := pcntr{reqs: make(chan uint64), resps: make(chan uint64)}


  go func() {
    //var cnt uint64 = 0
    for range pc.reqs {
      //cnt++
      counter++
      pc.resps <- counter//cnt
    }
  }()

  ch := make(chan uint64, 100)
  //a := pcntr(0)

  for i:= 1; i < 100 ; i++ {
    go func() {
      //u := uint64(i)
      ch <- pc.getNext()
      //if u != c {
        //fmt.Println("FALSE")
        //fmt.Printf("c = %d | u = %d \n", c, u)
      //}
      
    }()
   
  }

  max := uint64(0)
  for i:= 1; i < 100 ; i++ {
    m := <- ch
    if m > max {
      max = m
    }
  }

  fmt.Println("Done!")
  fmt.Println(counter)
  fmt.Println(max)

}

func (*nosync)getNext() uint64 {
  counter++
  return counter
}

func (*atom)getNext() uint64 {
  return atomic.AddUint64(&counter, 1)
}

func (*mutx)getNext() uint64 {
  cntMutex.Lock()
  counter++
  c := counter
  cntMutex.Unlock()
  return c
}

func (p *pcntr)getNext() uint64 {
  p.reqs <- 1
  return <- p.resps
}
