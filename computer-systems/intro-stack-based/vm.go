package vm

import (
	"fmt"
	"os"
)

const (
	//Pushaddr  = 0x01
	Store = 0x02
	Add   = 0x03
	Sub   = 0x04
	Halt  = 0xff
)

// Stretch goals

const (
	//Addi = 0x05
	//Subi = 0x06
	Jump = 0x07
	Beqz = 0x08
)

//Stack Based
const (
	Push  = 0x01
	Pushi = 0x05
	Pop   = 0x06
)

const (
	PC = iota
	Reg1
	Reg2
)

var InstrLens = [9]byte{0x01, 0x02, 0x02, 0x01, 0x01, 0x02, 0x01, 0x02, 0x02}

// Given a 256 byte array of "memory", run the stored program
// to completion, modifying the data in place to reflect the result
//
// The memory format is:
//
// 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f ... ff
// __ __ __ __ __ __ __ __ __ __ __ __ __ __ __ __ ... __
// ^==DATA===============^ ^==INSTRUCTIONS==============^
//
func compute(memory []byte) {
	pc := byte(8)
	var stack []byte
	lastIdx := -1
	// Keep looping, like a physical computer's clock
	for {

		op, arg := memory[pc], memory[pc+1]
		// decode and execute
		switch op {
		case Push:
			stack = append(stack, memory[arg])
			lastIdx++
		case Store:
			if lastIdx < 0 {
				fmt.Println("Nothing on the stack to store")
				os.Exit(1)
			}
			memory[arg] = stack[lastIdx]
			stack = stack[:lastIdx]
			lastIdx--
		case Add:
			if lastIdx < 1 {
				fmt.Println("Add requires at least two items on the stack")
				os.Exit(1)
			}
			stack[lastIdx-1] += stack[lastIdx]
			stack = stack[:lastIdx]
			lastIdx--
		case Sub:
			if lastIdx < 1 {
				fmt.Println("Sub requires at least two items on the stack")
				os.Exit(1)
			}
			stack[lastIdx-1] = stack[lastIdx] - stack[lastIdx-1]
			stack = stack[:lastIdx]
			lastIdx--
		case Pushi:
			stack = append(stack, arg)
			lastIdx++
		case Pop:
			if lastIdx < 0 {
				fmt.Println("Nothing on the stack to Pop")
				os.Exit(1)
			}
			stack = stack[:lastIdx]
			lastIdx--
		case Jump:
			pc = arg
			continue
		case Beqz:
			if lastIdx < 0 {
				fmt.Println("Nothing on the stack to check beqz")
				os.Exit(1)
			}
			if stack[lastIdx] == 0 {
				pc += arg
			}
		case Halt:
			return
		}
		//if next length 3 instruction extends past heap boundary, exit
		if pc == 255 || (pc == 254 && InstrLens[op] > 1) {
			fmt.Println("Instruction cut off  or falls off")
			return
		}
		pc += InstrLens[op]
	}
}
