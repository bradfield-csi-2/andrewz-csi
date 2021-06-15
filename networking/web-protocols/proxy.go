package main

import (
  "bytes"
  "encoding/binary"
  "errors"
  "fmt"
  "strings"
  "syscall"
)


//var pathsToCache map[string]bool
var pathToCache string
var cache map[string][]byte

func main() {
  pathToCache = "cache"
  cache = make(map[string][]byte)

  reqSock, err  := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
  if err != nil {
    fmt.Println(err)
  }
  err = syscall.SetsockoptInt(reqSock, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
  if err != nil {
    fmt.Println(err)
  }

  if err != nil {
    fmt.Println(err)
  }
 
 
  var sa syscall.SockaddrInet4
  sa.Port = 8000

  err = syscall.Bind(reqSock, &sa)
  if err != nil {
    fmt.Println(err)
  }

  err = syscall.Listen(reqSock, 0)
  if err != nil {
    panic(err)
  }

  for {
    acceptRecvFwdReply(reqSock)
  }
}
//end main

//Functions start
func acceptRecvFwdReply(reqSock int) {
  buf := make([]byte, 500)
  nfd, reqSockAddr, err := syscall.Accept(reqSock)
  if err != nil {
    panic(err)
  }

  fmt.Println("accepted")

  err = syscall.SetsockoptInt(nfd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
  if err != nil {
    panic(err)
  }

  fmt.Println("set resuse flag")

  //err = syscall.SetsockoptInt(nfd, syscall.SOL_SOCKET, syscall.SO_LINGER, 1)
  //if err != nil {
  //  panic(err)
  //}

  n, _, err := syscall.Recvfrom(nfd, buf, 0)
  if err != nil {
    panic(err)
  }

  fmt.Println("received from requester")

  shouldCache, path := checkIfCacheReqAndGetPathStr(buf[:n],pathToCache)
  fmt.Printf("should cache?: %v | reqPath: %s \n", shouldCache,path)
  //fmt.Println(reqFrom)
  //timeout := syscall.Timeval{Sec:10,Usec:0}
  //err = syscall.SetsockoptTimeval(nfd, syscall.SOL_SOCKET, syscall.SO_SNDTIMEO, &timeout)
  //if err != nil {
  //  fmt.Println(err)
  //}

  //err = syscall.SetsockoptTimeval(nfd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &timeout)

  var respBuf []byte
  if shouldCache {
    fmt.Println("using cache path")
    respBuf = getOrCacheResult(path, buf[:n])
  } else {
    fmt.Println("non cache path")
    respBuf = fwdReqAndGetResp(buf[:n])
    //hold for not cached
  }
 
  err = syscall.Sendto(nfd, respBuf, 0, reqSockAddr)
  if err != nil {
    panic(err)
  }
  fmt.Println("send response back to requester")
  n, _, err = syscall.Recvfrom(nfd, buf, 0)
  if err != nil {
    //panic(err)
    fmt.Println("timed out. did not receive response from requestser")
    fmt.Println(err)
    //fmt.Println("Did not receive response curl")
  } else {
    fmt.Println("received ack from requester")
    fmt.Printf("%d bytes: %s\n",n,string(buf[:n]))
  }
  syscall.Close(nfd)
   if err != nil {
    panic(err)
  }
} 


func getOrCacheResult(pathStr string, reqBuf []byte) []byte {
  respBuf, ok := cache[pathStr]
  if !ok {
    tempRslt := fwdReqAndGetResp(reqBuf)
    rsltToCache := make([]byte,len(tempRslt))
    copy(rsltToCache,tempRslt)
    cache[pathStr] = rsltToCache
    return rsltToCache
  }
  return respBuf
}

func fwdReqAndGetResp(reqBuf []byte) []byte{
  servSock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0) 
  err = syscall.SetsockoptInt(servSock, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
  if err != nil {
    panic(err)
  }

  //fmt.Println("set sersock flag")

  fwdSockAddr := syscall.SockaddrInet4{Port:9000, Addr: [4]byte{127,0,0,1}}
  err = syscall.Connect(servSock, &fwdSockAddr)
  if err != nil {
    panic(err)
  }
  fmt.Println("connected to py server")
  err = syscall.Sendto(servSock, reqBuf, 0, &fwdSockAddr)
  if err != nil {
    panic(err)
  }
  fmt.Println("forward request to pyserver")
  respBuf := make([]byte,500)
  n, _, err := syscall.Recvfrom(servSock, respBuf, 0)
  if err != nil {
    panic(err)
  }
  fmt.Println("receive response from py server")
  fmt.Printf("%d bytes: %s \n", n, respBuf[:n])
  fmt.Printf("FULL: %s \n", respBuf[:200])
  syscall.Close(servSock)
  fmt.Println("closed serv connection")
  //fmt.Println(servFrom)
  return respBuf[:n]

}


func checkIfCacheReqAndGetPathStr(reqBuf []byte, cachePath string) (bool, string) {
  pathStr, err := getPathFromRequest(reqBuf)
  if err != nil {
    panic(err)
  }
  return strings.HasPrefix(pathStr, cachePath), pathStr
}

func getPathFromRequest(reqBuf []byte) (string, error) {
  //foundfirst := false
  startIdx := 5//-1
  endIdx := -1
  for i:= startIdx; i < len(reqBuf) && i < 100; i++ {
    if reqBuf[i] == 0x20 {
        endIdx = i
        break
    }
  }
  if endIdx == -1 {
    return "", errors.New("http format error when parsting for path")
  }
  return strings.ToLower(string(reqBuf[startIdx: endIdx])), nil
}

func getHttpDataOffset(reqBuf []byte) (int, error) {
  var matchNum uint32
  matchBytes := []byte{0x0d,0x0a, 0x0d, 0x0a}
  matchBuf := bytes.NewReader(matchBytes)
  err := binary.Read(matchBuf, binary.BigEndian, &matchNum)
  //foundfirst := false
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }
  var checkByteNum uint32 = 0
  for i, b := range reqBuf {
    checkByteNum = (checkByteNum << 8) + uint32(b)
    if checkByteNum == matchNum {
      return i + 1, nil
    }
  }
  err = errors.New("match not found for http data offset")
  return -1, err 
}
