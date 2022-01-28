package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

const (
	NULL_MSG  = iota
	NODE_COMM //for things like adding/removing nodes, location, etc
	DATA_MSG  //for updating/storage/retrieval
)

const (
	REQUEST = iota
	RESPONSE
)

const (
	GET = iota
	SET
)

const DST_PORT = 1234
const MY_PORT = 8888

func main() {
	//TODO: create encoding scheme
	//1. have different types of messages
	//2. make them layered and peelable
	//2a. have a type and length to peel and pass?
	//2b. or just continue with a message schema?
	//3. need a scheme for requesting and responding - seems simple
	//4. Handle evolving schemas and types

	//use ports from start up?
	//have one client code
	fmt.Println("Started program")
	go handleSigInt()
	fieldMsgs()

}

func fieldMsgs() {

	for {
		test_receive()
	}
}

func test_receive() {
	recvSock, err := unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, 0)
	check(err)
	var rsockaddr unix.SockaddrInet6
	rsockaddr.Port = MY_PORT
	rsockaddr.Addr[15] = 1
	err = unix.Bind(recvSock, &rsockaddr)
	check(err)

	buf := make([]byte, 250)

	n, fromsockaddr, err := unix.Recvfrom(recvSock, buf, 0)
	check(err)
	//var fsockaddr unix.Sockaddr = &fromsockaddr
	fmt.Println("Received", string(buf[:n]))
	if f, ok := fromsockaddr.(*unix.SockaddrInet6); ok {
		fmt.Println(f.Port, f.Addr)
	} else {
		fmt.Println("Couldn't assert from as socckaddrinet6")
	}
	//fmt.Println(fromsockaddr.(*unix.SockaddrInet6))
	unix.Close(recvSock)

}

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

	//buf := make([]byte, 8)
	//st := 0
	for i := 0; i < 200; i++ {
		//put64(i, buf, &st)
		//err = unix.Send(sendSock, buf, 0)
		check(err)
		//st = 0
	}
	unix.Close(sendSock)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func handleSigInt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	fmt.Println("SIGINT Handler installed")
	s := <-c
	fmt.Println("Received", s, "...Exiting program")
	os.Exit(0)
}

func handleMessage(msg []byte) {
	if len(msg) == 0 {
		log.Println("Weird case handling 0 len messges")
		return //ignore
		//TODO: return message with error
	}
	msgTyp := msg[0]
	if msgTyp == NODE_COMM {
		//handle node communication
		log.Println("Node Comm not handled yet")
	} else if msgTyp == DATA_MSG {
		handleDataMsg(msg[1:])
	}
}

func handleDataMsg(msg []byte) {
	if len(msg) == 0 {
		log.Println("Weird case handling 0 len DATA messages")
		return //ignore
		//TODO: return message with error
	}
	dataMsgTyp := msg[0] //binary.BigEndian.Uint8(msg)

	switch dataMsgTyp {
	case REQUEST:
		handleDataRequest(msg[1:])
	case RESPONSE:
		log.Println("Data response not handled yet")
	default:
		log.Printf("DATA MSG TYPE: %d ... NOT HANDLED", dataMsgTyp)
	}
}

func handleDataRequest(msg []byte) {
	if len(msg) == 0 {
		log.Println("Weird case handling 0 len DATA Request")
		return //ignore
		//TODO: return message with error
	}
	dataCmd := msg[0] //binary.BigEndian.Uint8(msg)

	switch dataCmd {
	case GET:
		handleGetCommand(msg[1:])
	case SET:
		handleSetCommand(msg[1:])
	default:
		log.Printf("DATA COMMAND: %d ... NOT HANDLED", dataCmd)
	}
}

func handleDataResponse(msg []byte) {

}

func handleGetCommand(msg []byte) {
	//parse for key and response with message
	key := string(msg)

	val, ok := get(key)
	fmt.Println("GET KEY", key)
	if ok {
		fmt.Println(val)
	} else {
		fmt.Println("NOT FOUND")
	}

}

func get(key string) (string, bool) {
	return "", false
}

func handleSetCommand(msg []byte) {
	//parse for key and response with message
	i := 0
	keyLen := Get16toInt(msg, &i)
	key := GetString(msg, &i, keyLen)

	val := msg[i:]
	fmt.Println("SET KEY", key)
	fmt.Println("Val", val)
	/*
		if ok {
			fmt.Println(val)
		} else {
			fmt.Println("NOT FOUND")
		}
	*/

}

func GetString(buf []byte, i *int, len int) string {
	end := *i + int(len)
	retStr := string(buf[*i:end])
	*i = end
	return retStr
}

func Get16toInt(buf []byte, i *int) int {
	keyLen := binary.BigEndian.Uint16(buf)
	*i += 2
	return int(keyLen)

}
