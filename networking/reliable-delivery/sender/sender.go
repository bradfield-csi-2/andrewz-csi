package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	//"os"
	"golang.org/x/sys/unix"
	//"unix"
	//"time"
)

func main() {
	fmt.Println("hi i'm running")
	sendSock, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		fmt.Println("error with socket creation: ", err)
		panic(err)
	}

	var sendSockaddr unix.SockaddrInet4
	sendSockaddr.Port = 8888

	err = unix.Bind(sendSock, &sendSockaddr)

	if err != nil {
		fmt.Println("error with socket binding: ", err)
		panic(err)
	}

	dlen := 1024
	data := make([]byte, dlen)

	for i := 0; i < dlen; i++ {
		data[i] = 0x2a
	}

	sendData(sendSock, 55574, data)

	fmt.Println("END PROGRAM")
}

func sendData(csock, dstPort int, data []byte) {
	dstSockaddr := unix.SockaddrInet4{Port: dstPort, Addr: [4]byte{0, 0, 0, 0}}
	upperBound := (1 << 16)
	if len(data) >= upperBound {
		panic("too much data for send buffer")
	}
	var dataOff uint16 = 0
	hdrSize := 6
	maxPktSz := 100

	buf := make([]byte, 200)
	binary.BigEndian.PutUint64(buf, 0)

	unackedMap := make(map[uint16]uint16)

	//START Split and send inital packets
	for int(dataOff)+maxPktSz <= len(data) {
		binary.BigEndian.PutUint16(buf[4:], dataOff)
		unackedMap[dataOff] = dataOff + uint16(maxPktSz)
		writeBuf := bytes.NewBuffer(buf[:hdrSize])
		n, err := writeBuf.Write(data[int(dataOff) : int(dataOff)+maxPktSz]) //readBuf.WriteTo(writeBuf)
		if err != nil {
			panic(err)
		}
		if n != maxPktSz {
			panic("not 100?")
		}
		checksum := calcCheckSum(writeBuf.Bytes())

		//fmt.Println(string(writeBuf.String()))
		toSend := writeBuf.Bytes()
		binary.BigEndian.PutUint16(toSend[2:], checksum)
		err = unix.Sendto(csock, toSend, 0, &dstSockaddr)
		if err != nil {
			fmt.Println("error sending msg: ", err)
			panic(err)
		}
		dataOff += uint16(maxPktSz)
	}
	//send ending packet
	if int(dataOff) < len(data) {
		diff := len(data) - int(dataOff)
		binary.BigEndian.PutUint16(buf[4:], dataOff)
		unackedMap[dataOff] = dataOff + uint16(diff)
		writeBuf := bytes.NewBuffer(buf[:hdrSize])
		n, err := writeBuf.Write(data[int(dataOff):]) //readBuf.WriteTo(writeBuf)
		if err != nil {
			panic(err)
		}
		if int(n) != diff {
			panic("not equal to diff")
		}
		checksum := calcCheckSum(writeBuf.Bytes())

		//fmt.Println(string(writeBuf.String()))
		toSend := writeBuf.Bytes()
		binary.BigEndian.PutUint16(toSend[2:], checksum)
		err = unix.Sendto(csock, toSend, 0, &dstSockaddr)
		if err != nil {
			fmt.Println("error sending msg: ", err)
			panic(err)
		}
	}
	//END initial packet send

	rset, wset, eset := new(unix.FdSet), new(unix.FdSet), new(unix.FdSet)
	fmt.Printf(" sets: %v \n", rset)

	rset.Zero()
	wset.Zero()
	eset.Zero()

	timeout := unix.NsecToTimeval(1000)

	for len(unackedMap) > 0 {
		processAcks(10, csock, unackedMap, rset, wset, eset, timeout)

		spamUnackedPkts(unackedMap, csock, hdrSize, dstSockaddr, data)

	}

}

func processAcks(tryBeforeRespam int, csock int, unackedMap map[uint16]uint16, rset, wset, eset *unix.FdSet, timeout unix.Timeval) {
	buf := make([]byte, 200)
	for i := 0; i < tryBeforeRespam; i++ {
		rset.Set(csock)
		nReady, selErr := unix.Select(csock+1, rset, wset, eset, &timeout)
		if selErr != nil {
			fmt.Println("sel err: ", selErr)
			panic(selErr)
		}
		if nReady > 1 {
			panic("select found more than one")
		}
		if nReady == 1 {
			n, from, recvErr := unix.Recvfrom(csock, buf, 0)
			if recvErr != nil {
				panic(recvErr)
			}
			if n != 6 {
				panic("ack is not 6 bytes")
			}
			fmt.Printf("received from %v \n", from)
			//fmt.Println(string(buf[:n]))
			if buf[1] == 0x01 {
				seqNum := binary.BigEndian.Uint16(buf[4:])
				fmt.Printf("received ack for seqNum: %d \n", seqNum)
				delete(unackedMap, seqNum)
				fmt.Printf("map:  %v \n", unackedMap)
			} else {
				continue
			}
		} else {
			fmt.Println("breaking out, nothing ready")
			break
		}
	}
}

func spamUnackedPkts(unackedMap map[uint16]uint16, csock, hdrSize int, dstSockaddr unix.SockaddrInet4, data []byte) {
	buf := make([]byte, 200)
	for k, v := range unackedMap {
		binary.BigEndian.PutUint16(buf, 0)
		binary.BigEndian.PutUint16(buf[4:], k)
		writeBuf := bytes.NewBuffer(buf[:hdrSize])

		n, err := writeBuf.Write(data[int(k):int(v)])
		if err != nil {
			panic(err)
		}

		mapDiff := int(v) - int(k)
		if int(n) != mapDiff {
			panic("not map diff")
		}
		checksum := calcCheckSum(writeBuf.Bytes())

		//fmt.Println(string(writeBuf.String()))
		toSend := writeBuf.Bytes()
		binary.BigEndian.PutUint16(toSend[2:], checksum)

		err = unix.Sendto(csock, toSend, 0, &dstSockaddr)
		if err != nil {
			fmt.Println("error sending msg: ", err)
			panic(err)
		}
	}
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

func sendFinal(csock int, dstSockaddr unix.SockaddrInet4) {
	buf := make([]byte, 6)
	binary.BigEndian.PutUint16(buf, 2)
	binary.BigEndian.PutUint16(buf[4:], 0)
	//writeBuf := bytes.NewBuffer(buf[:hdrSize])
	err := unix.Sendto(csock, buf, 0, &dstSockaddr)
	if err != nil {
		fmt.Println("error sending msg: ", err)
		panic(err)
	}
}
