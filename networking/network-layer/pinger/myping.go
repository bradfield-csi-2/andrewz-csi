package main

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/sys/unix"
  "time"
)

const (
	Dropped = iota
	Echo
	Invalid
)

const (
	StartIdx     = 0
	IpPrtclIdx   = 9
	IpSrcAddrIdx = 12
	IpDstAddrIdx = 16
	IpStdDataIdx = 20
	IcmpTypeIdx  = 0
	IcmpDataIdx  = 8
)

const (
	ICMPProtocol  = 0x01
	TTLExceedType = 0x0b
	ICMPEchoType  = 0x00
)

type ipHdr struct {
	verHdrLen      byte
	ecn            byte
	dataLen        uint16
	id             [2]byte
	flagFragOffset [2]byte
	ttl            byte
	prtcl          byte
	checksum       [2]byte
	srcaddr        [4]byte
	dstaddr        [4]byte
	//options??
} //size == 20

type icmpHdr struct {
	typ      byte
	code     byte
	checksum [2]byte
	idnum    [2]byte
	seqnum   [2]byte
}

func main() {
	fmt.Println("hello")

	tsock, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_ICMP)
	if err != nil {
		panic(err)
	}

	var tsaddr unix.SockaddrInet4
	//tsaddr.Port = 8888

	err = unix.Bind(tsock, &tsaddr)

	if err != nil {
		fmt.Println("err with sock binding ", err)
		panic(err)
	}

	dstaddr := unix.SockaddrInet4{Addr: [4]byte{142, 250, 176, 206}}

	ping(tsock, 21, &dstaddr)

	fmt.Println("END PROGRAM")
}

func calcCheckSum(pkt []byte) uint16 {
	checksum := uint16(0)
	var i int
	for i = 0; i+1 < len(pkt); i += 2 {
		next := binary.BigEndian.Uint16(pkt[i:])
		temp := uint32(checksum) + uint32(next)
		carry := temp & (1 << 16)
		checksum += next
		if carry == (1 << 16) {
			checksum += 1
		}
	}
	if i < len(pkt) {
		next := uint16(pkt[i]) << 8
		temp := uint32(checksum) + uint32(next)
		carry := temp & (1 << 16)
		checksum += next
		if carry == (1 << 16) {
			checksum += 1
		}
	}
	return ^checksum
}

func ping(sock, ttl int, dstaddr *unix.SockaddrInet4) {
	err := unix.SetsockoptInt(sock, unix.IPPROTO_IP, unix.IP_TTL, ttl)

	if err != nil {
		fmt.Println("err with setting ttl ", err)
		panic(err)
	}

	pingPkt := createPingPkt(ttl, 2)

	err = unix.Sendto(sock, pingPkt, 0, dstaddr)

	if err != nil {
		panic(err)
	}

  resp, ok := tryGetResp(sock)
  if ok {
    fmt.Println(resp)
  }
	//fmt.Printf(" %v \n", recvBuf[:n])

	/*
	  for i, b := range recvBuf[:n] {
	    if i % 4 == 0 {
	      fmt.Println()
	    }
	    fmt.Printf(" %02x ", b)
	  }
	*/
	//
	//fmt.Println("<<>>>")

	//testType(recvBuf[:n])

}


func tryGetResp(sock int) ([]byte, bool) {
	for i := 0; i < 6; i++ {
		rset, wset, eset := new(unix.FdSet), new(unix.FdSet), new(unix.FdSet)
		timeout := unix.NsecToTimeval(5000000000)
		rset.Set(sock)
		nReady, selErr := unix.Select(sock+1, rset, wset, eset, &timeout)
		if selErr != nil {
			fmt.Println("sel err: ", selErr)
			panic(selErr)
		}
		if nReady > 1 {
			panic("select found more than one")
		}
		if nReady == 1 {

			recvBuf := make([]byte, 1500)
			n, from, recvErr := unix.Recvfrom(sock, recvBuf, 0)
			if recvErr != nil {
				fmt.Println("err receiving ", recvErr)
				panic(recvErr)
			}
			fmt.Printf("received from: %v \n", from)
      //testType(recvBuf[:n])
      //1 - get response type
      //if not valid then continue
      //else if dropped 
      //parse and match data
      //else if echo parse and match data?
      return recvBuf[:n], true
			break
		} else {
      time.Sleep(5 * time.Second)
      //wait five seconds?
			continue
		}
	}
  return make([]byte,0), false

}

func validatePingResp(pkt []byte) ok bool {
  //get type
  //if invalid return false
  //if dropped get src orig src and data. 
  //match data against expected
  //if matches then return ok
  //if echo get src and data
  //match data against expected, if ok then return ok
  //else return false
}

func hasEqualBytes(s1, s2 []byte) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}

func createPingPkt(ttl, i int) []byte {
	buf := make([]byte, 12)
	buf[0], buf[1], buf[2], buf[3] = 8, 0, 0, 0                     // type, code, checksum x2
	buf[4], buf[5], buf[6], buf[7] = 0, 0, 0, 0                     // id num x 2, seq num x2
	buf[8], buf[9], buf[10], buf[11] = 'a', 'z', byte(ttl), byte(i) // data/payload

	checksum := calcCheckSum(buf)
	binary.BigEndian.PutUint16(buf[2:], checksum)
	return buf
}

func getRespType(pkt []byte) (respType int, ok bool) {
	ipHdrLen := getIPHdrLen(pkt)
	prtcl := getIPDataProtocol(pkt)
	if prtcl != ICMPProtocol {
		fmt.Println("not icmpProtocl")
		fmt.Println(prtcl)
		return Invalid, false
	}
	//srcAddr := getIPSrcaddr(pkt)
	//dstAddr needed??
	icmpPkt := pkt[ipHdrLen:]
	icmpType := getICMPType(icmpPkt)
	if icmpType != TTLExceedType && icmpType != ICMPEchoType {
		fmt.Println("not icmptype")
		fmt.Println(icmpType)
		return Invalid, false
	} else if icmpType == TTLExceedType {
		respType = Dropped
	} else {
		respType = Echo
	}
	return respType, true
	// next ip header == icmpPkt[8:]
}

func getEchoSrcAndData(pkt []byte) (src, data []byte) {
	src = getIPSrcaddr(pkt)
	ipHdrLen := getIPHdrLen(pkt)
	icmpPkt := pkt[ipHdrLen:]
	data = icmpPkt[IcmpDataIdx:]
	return src, data
}

func getDroppedSrcAndData(pkt []byte) (dropSrc, origSrc, origData []byte) {
	dropSrc = getIPSrcaddr(pkt)
	dropHdrLen := getIPHdrLen(pkt)
	dropIcmpPkt := pkt[dropHdrLen:]
	pingReqPkt := dropIcmpPkt[IcmpDataIdx:]
	origSrc = getIPSrcaddr(pingReqPkt)
	pingReqHdrLen := getIPHdrLen(pingReqPkt)
	pingReqIcmpPkt := pingReqPkt[pingReqHdrLen:]
	origData = pingReqIcmpPkt[IcmpDataIdx:]
	return dropSrc, origSrc, origData
}

func testType(pkt []byte) {
	respType, ok := getRespType(pkt)
	fmt.Println(respType, ok)
}

func getIPHdrLen(pkt []byte) int {
	return int(pkt[StartIdx]&0x0f) * 4 //return in bytes
}

func getIPDataProtocol(pkt []byte) byte {
	return pkt[IpPrtclIdx]
}

func getIPSrcaddr(pkt []byte) []byte {
	return pkt[IpSrcAddrIdx:IpDstAddrIdx]
}

func getIPDstaddr(pkt []byte) []byte {
	return pkt[IpDstAddrIdx:IpStdDataIdx]
}

func getICMPType(pkt []byte) byte {
	return pkt[IcmpTypeIdx]
}
