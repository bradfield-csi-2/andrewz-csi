package main


import (
  //"bufio"
  "fmt"
  "os"
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

type recHdr struct {
  tsSec uint32 
  tsuSec uint32 
  inclLen uint32 
  origLen uint32 
  raw []byte 
}

type pRecord struct {
  secs uint32
  uSecs uint32
  inclLen uint32
  origLen uint32
  offset int64
  raw []byte
  hdr []byte
  data []byte
  eth ethPkt
}


type ethPkt struct {
  macDest [6]byte
  macSrc [6]byte
  ethType [2]byte
  payload []byte
  ipv4 ipv4Pkt
}

type ipv4Pkt struct {
  version uint8
  ihl uint8
  dscp uint8
  ecn uint8
  totLen uint16
  id uint16
  resFlag bool
  dfFlag bool
  mfFlag bool
  fragOffset uint16
  ttl uint8
  protocol uint8
  hdrChecksum uint16
  srcIP uint32
  destIP uint32
  options []byte
  hdr []byte
  data []byte
  tcp tcpPkt
}


type tcpPkt struct {
  srcPort uint16
  destPort uint16
  seqNum uint32
  ackNum uint32
  dataOffset uint8
  res uint8
  flags []string
  winSize uint16
  checksum uint16
  urgPtr uint16
  options []byte
  hdr []byte
  data []byte
  httpMessage string
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

  packets := make([]pRecord,0)

  for err == nil {
    n, err = raw.ReadAt(buff,off)

    p := getPRecord(buff)
    p.offset = off

    packets = append(packets, p)

    off += 16 + int64(p.inclLen)
  }

  //fmt.Println(err)

  //fmt.Println(len(packets))

  for i,pkt := range packets {
    //ethType := pkt.eth.ethType//pkt.data[12:14]

    //if ethType[0] != 0x08 || ethType[1] != 0x00 {
    //  fmt.Printf(">> %d: ???\n", i)
    //  fmt.Println(ethType)
    //}
    /*
    prtcl := pkt.eth.payload[9]
    if prtcl != 0x06 {
      fmt.Printf(">> %d: ???\n", i)
      fmt.Println(prtcl)
    }
    */

    //fmt.Printf(">> %02d :: macSrc: %v | srcIP: %v | srcPort: %v | macDest: %v | destIP: %v | destPort: %v | \n", i, pkt.eth.macSrc, pkt.eth.ipv4.srcIP, pkt.eth.ipv4.tcp.srcPort, pkt.eth.macDest, pkt.eth.ipv4.destIP, pkt.eth.ipv4.tcp.destPort)

    //fmt.Printf(">> %02d :: srcPort: %v | destPort %v |  tcpFlags : %v | seqNum: %v | ackNum: %v | data size: %v \n",i, pkt.eth.ipv4.tcp.srcPort, pkt.eth.ipv4.tcp.destPort, pkt.eth.ipv4.tcp.flags, pkt.eth.ipv4.tcp.seqNum, pkt.eth.ipv4.tcp.ackNum, len(pkt.eth.ipv4.tcp.data))
    if pkt.eth.ipv4.tcp.srcPort == 80 {
      fmt.Printf(">> %02d :: offset: %x | tcpFlags : %v | seqNum: %v | ackNum: %v | data size: %v \n", i, pkt.offset, pkt.eth.ipv4.tcp.flags, pkt.eth.ipv4.tcp.seqNum, pkt.eth.ipv4.tcp.ackNum, len(pkt.eth.ipv4.tcp.data))
      if i > 14 && i < 26 {
        /*
        for j, b := range pkt.raw {//reth.ipv4.tcp.data {
          fmt.Printf(" %02x", b)
          if (j + 1) % 16 == 0 {
            fmt.Println()
          }
        }
        */
        fmt.Println(pkt.eth.ipv4.tcp.httpMessage)
      }
    }
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


func getPRecord(raw []byte) pRecord {
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

  ethPktRaw := raw[16:(16 + inclLen)]
  ethPktStruct := getEthPkt(ethPktRaw)



  p := pRecord{secs: tsSec, uSecs: tsuSec, inclLen: inclLen, origLen: origLen, raw: raw[:(16 + inclLen)], hdr: raw[:16], data: raw[16:(16 + inclLen)], eth: ethPktStruct}
  return p
}


func getEthPkt(raw []byte) ethPkt {
  var dest [6]byte
  var src [6]byte
  var typ [2]byte
  
  for i,b := range raw[:6] {
    dest[i] = b
  }

  for i,b := range raw[6:12] {
    src[i] = b
  }

  for i,b := range raw[12:14] {
    typ[i] = b
  }
  
  ip := getIpPkt(raw[14:])

  e := ethPkt{macDest: dest, macSrc: src, ethType: typ, payload: raw[14:], ipv4: ip}
  return e
}

func getIpPkt(raw []byte) ipv4Pkt {
  version := uint8(raw[0]) >> 4
  ihl := uint8(raw[0]) & 0x0f
  dscp := uint8(raw[1]) >> 2
  ecn := uint8(raw[1]) & 0x03
  totLen := uint16(raw[2]) << 8 + uint16(raw[3])
  id := uint16(raw[4]) << 8 + uint16(raw[5])
  resFlg := (raw[6] & 0x80) != 0x00
  dfFlg := (raw[6] & 0x40) != 0x00
  mfFlg := (raw[6] & 0x20) != 0x00
  fragOff := (uint16(raw[6] & 0x1f) << 8 ) + uint16(raw[7])
  ttl := uint8(raw[8])
  protocol := uint8(raw[9])
  hdrChecksum := (uint16(raw[10]) << 8) + uint16(raw[11])
  srcIP := (uint32(raw[12]) << 24) + (uint32(raw[13]) << 16) + (uint32(raw[14]) << 8) + uint32(raw[15])
  destIP := (uint32(raw[16]) << 24) + (uint32(raw[17]) << 16) + (uint32(raw[18]) << 8) + uint32(raw[19])
  options := raw[20:(ihl * 4)]
  hdr := raw[:(ihl * 4)]
  data := raw[(ihl * 4):]
  tcp := getTcpPkt(data)

  ipPkt := ipv4Pkt{version: version,
    ihl: ihl,
    dscp: dscp,
    ecn: ecn,
    totLen: totLen,
    id: id,
    resFlag: resFlg,
    dfFlag: dfFlg,
    mfFlag: mfFlg,
    fragOffset: fragOff,
    ttl: ttl,
    protocol: protocol,
    hdrChecksum: hdrChecksum,
    srcIP: srcIP,
    destIP: destIP,
    options: options,
    hdr: hdr,
    data: data,
    tcp: tcp}
    
  return ipPkt
}

func getTcpPkt(raw []byte) tcpPkt {
  src := (uint16(raw[0]) << 8) + uint16(raw[1])
  dest := (uint16(raw[2]) << 8) + uint16(raw[3])
  seqNum := (uint32(raw[4]) << 24) + (uint32(raw[5]) << 16) + (uint32(raw[6]) << 8) + uint32(raw[7])
  ackNum := (uint32(raw[8]) << 24) + (uint32(raw[9]) << 16) + (uint32(raw[10]) << 8) + uint32(raw[11])
  dataOffset := uint8(raw[12]) >> 4
  flags := make([]string,0)

  if raw[12] & 0x01 != 0x0 {
    flags = append(flags,"ns")
  }

  if raw[13] & 0x80 != 0x00 {
    flags = append(flags,"cwr")
  }

  if raw[13] & 0x40 != 0x00 {
    flags = append(flags,"ece")
  }

  if raw[13] & 0x20 != 0x00 {
    flags = append(flags,"urg")
  }

  if raw[13] & 0x10 != 0x00 {
    flags = append(flags,"ack")
  }

  if raw[13] & 0x08 != 0x00 {
    flags = append(flags,"psh")
  }

  if raw[13] & 0x04 != 0x00 {
    flags = append(flags,"rst")
  }

  if raw[13] & 0x02 != 0x00 {
    flags = append(flags,"syn")
  }

  if raw[13] & 0x01 != 0x00 {
    flags = append(flags,"fin")
  }


  winsize := (uint16(raw[14]) << 8) + uint16(raw[15])
  checksum := (uint16(raw[16]) << 8) + uint16(raw[17])
  urgPtr := (uint16(raw[18]) << 8) + uint16(raw[19])
  options := raw[20:(dataOffset * 4)]
  hdr := raw[:(dataOffset * 4)]
  data := raw[(dataOffset * 4):]
  http := string(data)


  t := tcpPkt{srcPort: src, 
    destPort:dest, 
    seqNum: seqNum,
    ackNum: ackNum,
    dataOffset: dataOffset,
    flags: flags,
    winSize: winsize,
    checksum: checksum,
    urgPtr: urgPtr,
    options: options,
    hdr: hdr,
    data: data,
    httpMessage: http}

  return t
}


