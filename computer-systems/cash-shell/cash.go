package main

import (
	"bufio"
	//"bytes"
	"fmt"
	//"golang.org/x/sys/unix"
	"os"
	"strings"
	//"unsafe"
	"os/exec"
)

//1F4B8
const cashEmoji = "\xf0\x9f\x92\xb8"

func main() {
	fmt.Println("Welcome to ca$h shell")
	repl()
}

func repl() {
	fmt.Printf("ca$h %s  ", cashEmoji)
	//loop := true
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "exit" {
			break
		}
		//fmt.Println(text)
		eval(text)
		fmt.Printf("ca$h %s  ", cashEmoji)

	}
	fmt.Println("\nExiting")
}

func eval(input string) {
	args := strings.Split(input, " ")
	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else if len(args) == 1 {
		cmd = exec.Command(input)
	}
	b, err := cmd.Output()
	if err != nil {
		fmt.Errorf("%q \n", err)
	}
	fmt.Print(string(b))

	/*
		split := strings.Split(input, " ")
		cmd := split[0]
		if cmd == "echo" {
			//var b []bytes//bytes.Buffer
			output := strings.Join(split[1:], " ")
			oLen := len(output)
			b := make([]byte, oLen)
			unix.Syscall(unix.SYS_WRITE, uintptr(unsafe.Pointer(&unix.Stdout)), uintptr(unsafe.Pointer(&output)), uintptr(unsafe.Pointer(&oLen)))
			unix.Syscall(unix.SYS_READ, uintptr(unsafe.Pointer(&unix.Stdout)), uintptr(unsafe.Pointer(&b)), uintptr(unsafe.Pointer(&oLen)))
			fmt.Println(string(b))
		} else {
			fmt.Println(input)
		}
	*/
}
