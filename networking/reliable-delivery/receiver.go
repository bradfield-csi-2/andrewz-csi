package main

import (
	"encoding/binary"
	"fmt"
	"syscall"
	//"time"
)

func main() {
	fmt.Println("hi i'm receiving")
	lsock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		fmt.Println("error creating lsock: ", err)
		panic(err)
	}

	lsockaddr := syscall.SockaddrInet4{Port: 9999}
	err = syscall.Bind(lsock, &lsockaddr)
	if err != nil {
		fmt.Println("error binding lsock: ", err)
		panic(err)
	}

	buf := make([]byte, 500)
	//var n int
	//var from syscall.Sockaddr
	//sendBuf := make([]byte,64)

	//finalBuf := make([]byte, 2000)
	dstPort := 8888
	dstSockaddr := syscall.SockaddrInet4{Port: dstPort, Addr: [4]byte{0, 0, 0, 0}}
	for {
		n, from, recvErr := syscall.Recvfrom(lsock, buf, 0)
		if recvErr != nil {
			fmt.Println("receive error: ", recvErr)
			break
		}
		fmt.Printf("received from: %v\n", from)
		seqNum := binary.BigEndian.Uint16(buf[4:])
		fmt.Printf("Seq Num: %d\n", seqNum)
		fmt.Println(string(buf[4:n]))
		//recChecksum := binary.BigEndian.Uint16(buf)
		cChecksum := calcCheckSum(buf[:n])
		if cChecksum != 0 {
			fmt.Println("detected corruption: skip")
			continue
		}

		buf[0] = 0x00
		buf[1] = 0x01
		//writeBuf := bytes.NewBuffer(finalBuf[seqNum:])
		//writeBuf.Write(buf[4:n])

		err = syscall.Sendto(lsock, buf[:6], 0, &dstSockaddr)
		if err != nil {
			fmt.Println("error sending msg: ", err)
			panic(err)
		}
		fmt.Printf("should ack seqNum: %d \n", seqNum)
		//time.Sleep(10 * time.Millisecond)

	}

	err = syscall.Close(lsock)
	if err != nil {
		fmt.Println("error closing lsock: ", err)
		panic(err)
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
