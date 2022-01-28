package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	//"os"
)

const (
	GET = iota
	SET
)

const DiskFile = "./kv-disk-file"

type kvCmd struct {
	cmd      int
	key, val string
}

func main() {
	fmt.Println("Started program")
	go handleSigInt()
	setupFile()
	repl()
}

func setupFile() {
	f, err := os.Open(DiskFile)
	defer f.Close()
	if os.IsNotExist(err) {
		f, err = os.Create(DiskFile)
		if err != nil {
			panic(err)
		}
		m := make(map[string]string)
		var buf []byte
		buf, err = json.Marshal(m)
		f.Write(buf)
	}
	if err != nil {
		panic(err)
	}
}

func repl() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()
		inCmd := scanner.Text()
		cmd, err := parseCommand(inCmd)
		if err != nil {
			fmt.Println(err)
			commandHelper()
		} else {
			executeCmd(cmd)
		}

	}
}

func executeCmd(cmd kvCmd) {
	//cmd struct
	switch cmd.cmd {
	case GET:
		if val, ok := get(cmd.key); ok {
			fmt.Println("GET: ", cmd.key, " -> ", val)
		} else {
			fmt.Println("KEY NOT FOUND")
		}
	case SET:
		if ok := set(cmd.key, cmd.val); ok {
			fmt.Println("SET OK")
		} else {
			fmt.Println("Could not set")
		}
	default:
		fmt.Println("Command: ", cmd.cmd)
		panic("Command not handled")
	}
}

func get(key string) (string, bool) {
	f, err := os.OpenFile(DiskFile, 0, 0)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	info, err := f.Stat()
	if err != nil {
		panic(err)
	}
	buf := make([]byte, info.Size())
	_, err = f.Read(buf)
	if err != nil {
		panic(err)
	}
	var m map[string]string
	err = json.Unmarshal(buf, &m)
	if err != nil {
		panic(err)
	}

	val, ok := m[key]
	return val, ok
}

func set(key, value string) bool {
	f, err := os.OpenFile(DiskFile, os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	info, err := f.Stat()
	if err != nil {
		panic(err)
	}
	buf := make([]byte, info.Size())
	_, err = f.Read(buf)
	if err != nil {
		panic(err)
	}
	var m map[string]string
	json.Unmarshal(buf, &m)
	//f.Close()
	m[key] = value
	//err = f.Truncate(0)
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	buf, err = json.Marshal(m)
	if err != nil {
		panic(err)
	}
	n, err := f.Write(buf)
	if err != nil {
		panic(err)
	}
	err = f.Truncate(int64(n))
	if err != nil {
		panic(err)
	}
	//return val, ok
	return true
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

func commandHelper() {
	fmt.Println("[get|set] [key|key=value]")
}

func parseCommand(inputCmd string) (kvCmd, error) {
	trimmed := strings.TrimSpace(inputCmd)
	tokens := strings.Split(trimmed, " ")
	cmd := kvCmd{0, "", ""}
	if len(tokens) != 2 || strings.Contains(tokens[1], " ") {
		//fmt.Println("Must have one and only one space.")
		return cmd, errors.New("Command format error: Must have only one space")
		//os.Exit(0)
	}
	first := strings.ToLower(tokens[0])
	if first != "get" && first != "set" {
		return cmd, errors.New("Command not recognized")
	}
	cmd.cmd = GET
	cmd.key = tokens[1]
	if first == "set" {
		cmd.cmd = SET
		kvparam := strings.Split(tokens[1], "=")
		if len(kvparam) != 2 {
			return cmd, errors.New("Command format error: Set must have = separating key-value pair")
		}

		cmd.key = kvparam[0]
		cmd.val = kvparam[1]
	}

	return cmd, nil

}
