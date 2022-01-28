package main

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {
	fmt.Println("hi")
	test_send()
}

const DST_PORT = 1234
const MY_PORT = 8888

func test_send() {
	sendSock, err := unix.Socket(unix.AF_INET6, unix.SOCK_STREAM, 0)
	check(err)

	var sendSockAddr unix.SockaddrInet6
	sendSockAddr.Port = MY_PORT
	sendSockAddr.Addr[15] = 1

	//err = unix.Bind(sendSock, &sendSockAddr)
	var dstAddr unix.SockaddrInet6
	dstAddr.Port = DST_PORT
	dstAddr.Addr[15] = 1
	err = unix.Connect(sendSock, &dstAddr)

	check(err)

	buf := make([]byte, 8)
	st := 0
	for i := 0; i < 200; i++ {
		put64(i, buf, &st)
		err = unix.Send(sendSock, buf, 0)
		check(err)
		st = 0
	}
	unix.Close(sendSock)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func put64(u int, buf []byte, i *int) {
	binary.BigEndian.PutUint64(buf[*i:], uint64(u))
	*i += 8
}
