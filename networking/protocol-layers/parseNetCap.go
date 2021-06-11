package main


import (
  //"bufio"
  "bytes"
  "encoding/binary"
  "io"
  "fmt"
  "os"
  "sort"
)


type globHdr struct {
  magicnum uint32   
  verMajor uint16 
  verMinor uint16 
  zone int32 
  sigfigs uint32 
  snaplen uint32 
  network uint32 
  raw []byte 
}

type pktRec struct {
  secs uint32
  uSecs uint32
  inclLen uint32
  origLen uint32
  capFile *os.File
  fileOffset int64
  dataOffset uint32
  ethFrameOffset uint32
  ipv4FrameOffset uint32
  tcpFrameOffset uint32
  httpMsgOffset uint32
}



func main() {
  raw, err := os.Open("./net.cap")
  defer raw.Close()
  if err != nil {
    //bufio.NewReader(raw)
    fmt.Println(err)
  }

  buff := make([]byte,500)

  var off int64 = 0

  n, err := raw.ReadAt(buff,off)
  //fmt.Printf(">>>>>> Read %d bytes at offset %d \n", n, off)
  //fmt.Println("Global Packet Header")
  gHdr := getGlobHdr(buff)
  off = 24


  buff = make([]byte, 16 + gHdr.snaplen)

  packets := make([]pktRec,0)

  var p pktRec

  for err == nil {
    n, err = raw.ReadAt(buff,off)

    p = parsePRecord(buff, raw, off)

    pktBin := buff[:p.getTotLen()]

    tcpFrame := p.getTcpFrame(pktBin)

    srcPort,_ := getTcpPorts(tcpFrame)
    
    if srcPort == 80 {
      packets = append(packets, p)
    }
    off += 16 + int64(p.inclLen)
  }

  if err == io.EOF {
    localOff := 16 + p.inclLen
    for int(localOff) < n {
      endBuff := buff[localOff:]
      p = parsePRecord(endBuff, raw, off)
      pktBin := endBuff[:p.getTotLen()]

      tcpFrame := p.getTcpFrame(pktBin)

      srcPort,_ := getTcpPorts(tcpFrame)
    
      if srcPort == 80 {
        packets = append(packets, p)
      }
      off += 16 + int64(p.inclLen)
      localOff += 16 + p.inclLen
    }
    if int(localOff) != n {
      fmt.Println(localOff)
      fmt.Println(n)
      panic("should be equal")
    }
  } else {
    fmt.Println(err)
  }

  sort.Slice(packets, 
    func(i, j int) bool { 
      pktBinI := packets[i].getRaw(buff)
      tcpFrameI := packets[i].getTcpFrame(pktBinI)
      seqNumI,_ := getTcpSeqAckNum(tcpFrameI)
      pktBinJ := packets[j].getRaw(buff)
      tcpFrameJ := packets[j].getTcpFrame(pktBinJ)
      seqNumJ,_ := getTcpSeqAckNum(tcpFrameJ)
 
      return seqNumI < seqNumJ 
    })

  uniqSortedPkts := make([]pktRec,0)

  var lastSeqNum uint32 = 0 
  for _, pkt := range packets {
    currSeqNum,_ := getTcpSeqAckNum(pkt.getTcpFrame(pkt.getRaw(buff)))
    if currSeqNum != lastSeqNum {
      uniqSortedPkts = append(uniqSortedPkts, pkt)
      lastSeqNum = currSeqNum
    }
  }

  httpResponse := make([]byte,0)

  for i,pkt := range uniqSortedPkts { 
    if i == 0 || i == 41 {
      continue
    }
    //pkt.printPktRec(i, buff)
    httpRespBytes := pkt.getHttpMsg(pkt.getRaw(buff))
    httpResponse = append(httpResponse, httpRespBytes...)
  }


  httpDataOffset := findHttpResponseDataOffset(httpResponse)


  out, err := os.Create("./flag.jpg")
  defer out.Close()
  if err != nil {
    panic(err)
  }


  _, err = out.Write(httpResponse[httpDataOffset:])
  if err != nil {
    panic(err)
  }
}

func findHttpResponseDataOffset(msgBytes []byte) uint32 {
 var matchNum uint32
 matchBytes := []byte{0x0d,0x0a, 0x0d, 0x0a}
 buf := bytes.NewReader(matchBytes)
 err := binary.Read(buf, binary.BigEndian, &matchNum)
 if err != nil {
   fmt.Println("binary.Read failed:", err)
 }
 var checkByteNum uint32 = 0
 for i, b := range msgBytes {
   checkByteNum = (checkByteNum << 8) + uint32(b)
   if checkByteNum == matchNum {
     return uint32(i) + 1
   }
 }
 fmt.Println("match not found for http data offset")
 return 0

}

func getGlobHdr(raw []byte) globHdr {
  var magicnum uint32
  var verMjr uint16
  var verMnr uint16
  var zone int32
  var sigfigs uint32
  var snaplen uint32
  var network uint32
  
  buf := bytes.NewReader(raw[:4])
  err := binary.Read(buf, binary.BigEndian, &magicnum)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  buf = bytes.NewReader(raw[4:6])
  err = binary.Read(buf, binary.BigEndian, &verMjr)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  buf = bytes.NewReader(raw[4:6])
  err = binary.Read(buf, binary.BigEndian, &verMnr)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }
  buf = bytes.NewReader(raw[8:12])
  err = binary.Read(buf, binary.BigEndian, &zone)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  buf = bytes.NewReader(raw[12:16])
  err = binary.Read(buf, binary.BigEndian, &sigfigs)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  buf = bytes.NewReader(raw[16:20])
  err = binary.Read(buf, binary.BigEndian, &snaplen)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  buf = bytes.NewReader(raw[20:24])
  err = binary.Read(buf, binary.BigEndian, &network)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  g := globHdr{magicnum: uint32(magicnum),
    verMajor: verMjr, 
    verMinor: verMnr,
    zone: zone,
    sigfigs: sigfigs,
    snaplen: snaplen,
    network: network,
    raw: raw[:24]}
  return g
}


func parsePRecord(raw []byte, capFile *os.File, filOff int64) pktRec {
  var tsSec uint32
  var tsuSec uint32
  var inclLen uint32
  var origLen uint32
  
  buf := bytes.NewReader(raw[:4])
  err := binary.Read(buf, binary.LittleEndian, &tsSec)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  buf = bytes.NewReader(raw[4:8])
  err = binary.Read(buf, binary.LittleEndian, &tsuSec)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  buf = bytes.NewReader(raw[8:12])
  err = binary.Read(buf, binary.LittleEndian, &inclLen)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  buf = bytes.NewReader(raw[12:16])
  err = binary.Read(buf, binary.LittleEndian, &origLen)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  dataOffset := uint32(16)
  endOfPktRec := dataOffset + inclLen
  ethFrameOffset := dataOffset
  ipv4FrameOffset := ethFrameOffset + getEthFramePayloadOffset(raw[ethFrameOffset:endOfPktRec])
  tcpFrameOffset := ipv4FrameOffset + getIpFrameHdrLen(raw[ipv4FrameOffset:endOfPktRec])
  httpMsgOffset := tcpFrameOffset + getTcpDataOffset(raw[tcpFrameOffset:endOfPktRec])

  p := pktRec{secs: tsSec,
    uSecs: tsuSec,
    inclLen: inclLen,
    origLen: origLen,
    capFile: capFile,
    fileOffset: filOff,
    dataOffset: dataOffset,
    ethFrameOffset: ethFrameOffset,
    ipv4FrameOffset: ipv4FrameOffset,
    tcpFrameOffset: tcpFrameOffset,
    httpMsgOffset: httpMsgOffset}
  return p
}


func getEthFramePayloadOffset(rawEthFrame []byte) uint32 {
  return 14
}


func getIpFrameHdrLen(rawIPFrame []byte) uint32 {
  ihl := uint8(rawIPFrame[0]) & 0x0f
  return uint32(ihl) * 4
}


func getTcpDataOffset(rawTcpFrame []byte) uint32 {
  dataOffset := uint32(rawTcpFrame[12]) >> 4
  dataOffset *= 4
  return dataOffset
}

func (p *pktRec)getTotLen() uint32 {
  return 16 + p.inclLen
}

func (p *pktRec)getRaw(buf []byte) []byte {
  n, err := p.capFile.ReadAt(buf,p.fileOffset)
  if uint32(n) < p.getTotLen() {
    panic("read bytes can't be less than packet length")
  }
  if err != nil && err != io.EOF {
    panic(err)
  }
  return buf[:p.getTotLen()]
}

func (p *pktRec)getEthFrame(raw []byte) []byte {
  return raw[16:p.getTotLen()]
}

func (p *pktRec)getIpFrame(raw []byte) []byte {
  return raw[p.ipv4FrameOffset:]
}

func (p *pktRec)getTcpFrame(raw []byte) []byte {
  return raw[p.tcpFrameOffset:]
}


func (p *pktRec)getHttpMsg(raw []byte) []byte {
  return raw[p.httpMsgOffset:]
}

func (p *pktRec) printPktRec(i int, buf []byte) {
  raw := p.getRaw(buf)
  tcpFrame := p.getTcpFrame(raw)
  src, dest := getTcpPorts(tcpFrame)
  seq, ack := getTcpSeqAckNum(tcpFrame)
  flags := getTcpFlags(tcpFrame)
  httpMsgBytes := p.getHttpMsg(raw)
  fmt.Printf(">> %02d :: src port: %06d | dest port: %06d | seq num: %010d | ack num: %010d | datalen: %v | tcp flags: %v \n", i, src, dest, seq, ack, len(httpMsgBytes), flags)
}


func getTcpSeqAckNum(tcpFrame []byte) (seqNum, ackNum uint32) {
  buf := bytes.NewReader(tcpFrame[4:8])
  err := binary.Read(buf, binary.BigEndian, &seqNum)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  buf = bytes.NewReader(tcpFrame[8:12])
  err = binary.Read(buf, binary.BigEndian, &ackNum)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }


  return
} 

func getTcpPorts(tcpFrame []byte) (src, dest uint16) {
  buf := bytes.NewReader(tcpFrame[:2])
  err := binary.Read(buf, binary.BigEndian, &src)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  buf = bytes.NewReader(tcpFrame[2:4])
  err = binary.Read(buf, binary.BigEndian, &dest)
  if err != nil {
    fmt.Println("binary.Read failed:", err)
  }

  return src, dest
} 


func getTcpFlags(tcpFrame []byte) []string {
  flags := make([]string,0)

  if tcpFrame[12] & 0x01 != 0x0 {
    flags = append(flags,"ns")
  }

  if tcpFrame[13] & 0x80 != 0x00 {
    flags = append(flags,"cwr")
  }

  if tcpFrame[13] & 0x40 != 0x00 {
    flags = append(flags,"ece")
  }

  if tcpFrame[13] & 0x20 != 0x00 {
    flags = append(flags,"urg")
  }

  if tcpFrame[13] & 0x10 != 0x00 {
    flags = append(flags,"ack")
  }

  if tcpFrame[13] & 0x08 != 0x00 {
    flags = append(flags,"psh")
  }

  if tcpFrame[13] & 0x04 != 0x00 {
    flags = append(flags,"rst")
  }

  if tcpFrame[13] & 0x02 != 0x00 {
    flags = append(flags,"syn")
  }

  if tcpFrame[13] & 0x01 != 0x00 {
    flags = append(flags,"fin")
  }
  
  return flags
}
