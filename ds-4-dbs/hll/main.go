package main

import (
	"bufio"
	"fmt"
  "hash/fnv"
  "sort"
  //"unsafe"
	//"log"
	//"math/rand"
  "math"
  "math/bits"
	"os"
	//"time"
)

const (
	path     = "/usr/share/dict/words"
)

type nums []uint8

func (n nums) Len() int           { return len(n) }
func (n nums) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n nums) Less(i, j int) bool { return n[i] < n[j] }


func loadWords() { //limit int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		//return nil, err
    panic(err)
	}
	//var result []string
  var n int
	s := bufio.NewScanner(f)
	for s.Scan() {//&& (limit == 0 || len(result) < limit) {
		//result = append(result, s.Text())
    //fmt.Println(s.Text())
    n++
	}
  fmt.Println(n)
	if err := s.Err(); err != nil {
		//return nil, err
    panic(err)
	}
	//return result, nil
}

func maxLeadingZeros() int {
  fnvHash := fnv.New32()
	f, err := os.Open(path)
	if err != nil {
    panic(err)
	}
  var max int = 0
	s := bufio.NewScanner(f)
	for s.Scan() {
    //n++
    //iword := s.Text()
    fnvHash.Write(s.Bytes())
    fnvSum := fnvHash.Sum32()
    //fnvHash.Reset()
    tz := bits.TrailingZeros32(fnvSum)
    if tz > max {
      max = tz
    }
	}
	if err := s.Err(); err != nil {
    panic(err)
	}
  return max
}	


func loglogCard() float64 {
  regs := make([]uint8, 2048)
  fnvHash := fnv.New32()
	f, err := os.Open(path)
	if err != nil {
    panic(err)
	}
	s := bufio.NewScanner(f)
	for s.Scan() {
    //n++
    //iword := s.Text()
    fnvHash.Write(s.Bytes())
    fnvSum := fnvHash.Sum32()
    //fnvHash.Reset()
    idx := fnvSum >> (32 - 11)
    tz := uint8(bits.TrailingZeros32(fnvSum))
    if tz > regs[idx] {
      regs[idx] = tz
    }
	}
	if err := s.Err(); err != nil {
    panic(err)
	}
  var sum uint32 = 0
  for _, r := range regs{
    sum += uint32(r)
  }
  avg := float64(sum)  /  float64(2048)

  return float64(0.79402) * float64(2048) * math.Pow(2, avg)
 
}

func hllCard() float64 {
  regs := make([]uint8, 2048)
  fnvHash := fnv.New64()
	f, err := os.Open(path)
	if err != nil {
    panic(err)
	}
	s := bufio.NewScanner(f)
	for s.Scan() {
    //n++
    //iword := s.Text()
    fnvHash.Write(s.Bytes())
    fnvSum := fnvHash.Sum64()
    //fnvHash.Reset()
    idx := fnvSum >> (64 - 11)
    tz := uint8(bits.TrailingZeros64(fnvSum))
    if tz > regs[idx] {
      regs[idx] = tz
    }
	}
	if err := s.Err(); err != nil {
    panic(err)
	}
  sort.Sort(nums(regs))
  //fmt.Println(regs)
  //fmt.Println(regs[0])
  //fmt.Println(regs[2047])
  //regs = regs[300:1748]
  var sum uint32 = 0
  for _, r := range regs{
    sum += uint32(r)
  }
  avg := float64(sum)  /  float64(2048)

  return float64(0.79402) * float64(2048) * math.Pow(2, avg)
 
}

func main() {
  //loadWords()
  card := hllCard()//loglogCard()
  fmt.Println(card)
}
