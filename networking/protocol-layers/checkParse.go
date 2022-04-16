package main


import (
  //"bufio"
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
  fmt.Println("hello")

  raw, err := os.Open("./net.cap")
  defer raw.Close()
  if err != nil {
    //bufio.NewReader(raw)
    fmt.Println(err)
  }

  buff := make([]byte,500)

  var off int64 = 0

  n, err := raw.ReadAt(buff,off)
  fmt.Printf(">>>>>> Read %d bytes at offset %d \n", n, off)
  fmt.Println("Global Packet Header")
  //for _,b := range buff {
  //  fmt.Printf("%02x ", b)
  //}
  gHdr := getGlobHdr(buff)
  off = 24


  buff = make([]byte, 16 + gHdr.snaplen)

  packets := make([]pktRec,0)

  //totIncLen := 0

  var p pktRec
  for err == nil {
    n, err = raw.ReadAt(buff,off)

    //p := getPRecord(buff)
    //p.offset = off
    p = parsePRecord(buff, raw, off)

    
    packets = append(packets, p)
    off += 16 + int64(p.inclLen)
    //totIncLen += int(p.inclLen)
  }

  if err == io.EOF {
    localOff := 16 + p.inclLen
    for int(localOff) < n {
      endBuff := buff[localOff:]
      p = parsePRecord(endBuff, raw, off)
      packets = append(packets,p)
      off += 16 + int64(p.inclLen)
      //totIncLen += int(p.inclLen)
      localOff += 16 + p.inclLen
    }
  }

  if err != io.EOF {
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

  //fmt.Println(err)
  //for i, pkt := range packets {
    //pkt.printPktRec(i,buff)
  //}
  packets[7].checkPkt(buff)

  //hdrLen := 16 * len(packets) + 24

  //totalCalcLen := totIncLen + hdrLen

  //fmt.Printf(" final off = %d || hdrLens = %d || dataLens = %d || totalCalcLen = %d \n", off, hdrLen, totIncLen, totalCalcLen)

}

func (p *pktRec) checkPkt( buf []byte) {
  raw := p.getRaw(buf)
  //ethFrame := p.getEthFrame(raw)
  //ipFrame := p.getIpFrame(raw)
  //headerLen := uint(ipFrame[0] & 0x0f)
  //totLen := uint(ipFrame[2]) << 8 + uint(ipFrame[3])

  tcpFrame := p.getTcpFrame(raw)
  src, dest := getTcpPorts(tcpFrame)
  seq, ack := getTcpSeqAckNum(tcpFrame)
  flags := getTcpFlags(tcpFrame)
  httpMsgBytes := p.getHttpMsg(raw)
  fmt.Printf(">> :: src port: %06d | dest port: %06d | seq num: %010d | ack num: %010d | datalen: %v | tcp flags: %v \n", src, dest, seq, ack, len(httpMsgBytes), flags)
  fmt.Println(p.httpMsgOffset)
  
  for j,b := range httpMsgBytes {
    if j % 4 == 0 {
      fmt.Println()
    }
    fmt.Printf(" %02x", b)
  }
  

}


func getGlobHdr(raw []byte) globHdr {
  magicRaw := raw[:4]
  magicnum := 0
  for i,m := range magicRaw {
    magicnum += int(m) << (8 * i)
  }

  verMjrRaw := raw[4:6]
  verMnrRaw := raw[6:8]

  verMjr := 0
  verMnr := 0

  for i,v := range verMjrRaw {
    verMjr += int(v) << (8 * i)
  }

  for i,v := range verMnrRaw {
    verMnr += int(v) << (8 * i)
  }

  zone := 0

  for i,z := range raw[8:12] {
    zone += int(z) << (8 * i)
  }

  var sigfigs uint32 = 0

  for i,s := range raw[12:16] {
    sigfigs += uint32(s) << (8 * i)
  }


  var snaplen uint32 = 0

  for i,s := range raw[16:20] {
    snaplen += uint32(s) << (8 * i)
  }

  var network uint32 = 0

  for i,s := range raw[16:24] {
    network += uint32(s) << (8 * i)
  }




  //fmt.Println(magicnum)
  g := globHdr{magicnum: uint32(magicnum),
    verMajor: uint16(verMjr), 
    verMinor: uint16(verMnr),
    zone: int32(zone),
    sigfigs: sigfigs,
    snaplen: snaplen,
    network: network,
    raw: raw[:24]}
  return g


}


func parsePRecord(raw []byte, capFile *os.File, filOff int64) pktRec {
  var tsSec uint32 = 0
  
  for i,s := range raw[:4] {
    tsSec += uint32(s) << (8 * i)
  }

  var tsuSec uint32 = 0

  for i,u := range raw[4:8] {
    tsuSec += uint32(u) << (8 * i)
  }

  var inclLen uint32 = 0

  for i, l := range raw[8:12] {
    inclLen += uint32(l) << (8 * i)
  }

  var origLen uint32 = 0

  for i,l := range raw[12:16] {
    origLen += uint32(l) << (8 * i)
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
  //ethFrame := p.getEthFrame(raw)
  //ipFrame := p.getIpFrame(raw)
  //headerLen := uint(ipFrame[0] & 0x0f)
  //totLen := uint(ipFrame[2]) << 8 + uint(ipFrame[3])

  tcpFrame := p.getTcpFrame(raw)
  src, dest := getTcpPorts(tcpFrame)
  seq, ack := getTcpSeqAckNum(tcpFrame)
  flags := getTcpFlags(tcpFrame)
  httpMsgBytes := p.getHttpMsg(raw)
  fmt.Printf(">> %02d :: src port: %06d | dest port: %06d | seq num: %010d | ack num: %010d | datalen: %v | tcp flags: %v \n", i, src, dest, seq, ack, len(httpMsgBytes), flags)
  //fmt.Printf(">> %02d :: ts: %10d | ts micro: %10d | incl len: %5d | origlen: %5d \n", i, p.secs, p.uSecs, p.inclLen, p.origLen)
  //fmt.Printf(">> %02d :: protocol: %v | dest mac: %v | src mac: %v \n", i, ethFrame[12:14], ethFrame[:6], ethFrame[6:12])
  //fmt.Printf(">> %02d :: headerLen : %v | totLen in bytes: %v | rawLen in bytes | %v | endBytes: %v  \n", i, headerLen, totLen, len(ipFrame), ipFrame[len(ipFrame) - 2:])


  /*
  httpMsgBytes := p.getHttpMsg(raw)
  fmt.Printf(">>> %02d \n", i)
  fmt.Println(string(httpMsgBytes))
  fmt.Printf("<<< %02d \n", i)
  */
}


func getTcpSeqAckNum(tcpFrame []byte) (seqNum, ackNum uint32) {

  seqNum = (uint32(tcpFrame[4]) << 24) + (uint32(tcpFrame[5]) << 16) + (uint32(tcpFrame[6]) << 8) + uint32(tcpFrame[7])
  ackNum = (uint32(tcpFrame[8]) << 24) + (uint32(tcpFrame[9]) << 16) + (uint32(tcpFrame[10]) << 8) + uint32(tcpFrame[11])
  return
} 

func getTcpPorts(tcpFrame []byte) (src, dest uint16) {
  src = (uint16(tcpFrame[0]) << 8) + uint16(tcpFrame[1])
  dest = (uint16(tcpFrame[2]) << 8) + uint16(tcpFrame[3])
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
