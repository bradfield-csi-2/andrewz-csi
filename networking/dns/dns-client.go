package main

import (
  //"bytes"
  "encoding/binary"
  "fmt"
  "golang.org/x/sys/unix"
  "strings"
  //"time"
)

func main() {
  fmt.Println("hello")
  pid := unix.Getpid()
  fmt.Println(pid)
  socket, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
  if err != nil {
    fmt.Println(err)
  }
  //sa := unix.SockaddrInet4{}
  var sa unix.SockaddrInet4
  fmt.Println(sa)

  err = unix.Bind(socket, &sa)
  if err != nil {
    fmt.Println(err)
  }

  fmt.Println(sa)


  //dnsRawSockAddr := unix.RawSockaddrInet4{Family: unix.AF_INET, Port: 53, Addr: [4]byte{8,8,8,8}}
  //fmt.Println(dnsRawSockAddr)
  buf := make([]byte,6000)
  dnsSockAddr := unix.SockaddrInet4{Port:53, Addr: [4]byte{8,8,8,8}}

  //var flgOne byte = 0x00
  //var flgTwo byte = 0x80
  //qdCount := uint16(0)

  id := uint16(1)
  flags := uint16(0x0100)
  qdCount := uint16(1)
  anCount := uint16(0)
  nsCount := uint16(0)
  arCount := uint16(0)

  domainName := "google.com"

  parts := strings.Split(domainName,".")

  qType := uint16(1)
  qClass := uint16(1)



  binary.BigEndian.PutUint16(buf,id)
  binary.BigEndian.PutUint16(buf[2:],flags)
  binary.BigEndian.PutUint16(buf[4:],qdCount)
  binary.BigEndian.PutUint16(buf[6:],anCount)
  binary.BigEndian.PutUint16(buf[8:],nsCount)
  binary.BigEndian.PutUint16(buf[10:],arCount)

  k := 12
  for _,str := range parts {
    buf[k] = byte(len(str))
    k++
    for i := 0; i < len(str); i++ {
      buf[k] = str[i]
      k++
    }
  }
  buf[k] = 0x00
  k++
  binary.BigEndian.PutUint16(buf[k:],qType)
  binary.BigEndian.PutUint16(buf[k+2:],qClass)



  unix.Sendto(socket, buf[:100], 0, &dnsSockAddr)
  //time.Sleep(10)
  n, from, err := unix.Recvfrom(socket, buf, 0)
  //Recvfrom(fd int, p []byte, flags int) (n int, from Sockaddr, err error)
  if err != nil {
    fmt.Println(err)
  }
  fmt.Println(from)
  //fmt.Println(buf[:n])
  for i,b := range buf[:n] {
    if i % 16 == 0 {
      fmt.Println()
    }
    fmt.Printf(" %02x",b)
  }


  //fmt.Println(socket)
}


