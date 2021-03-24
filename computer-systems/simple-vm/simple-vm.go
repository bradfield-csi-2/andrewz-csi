package main

import (
	"errors"
	"fmt"
)

const (
	PC = iota
	Reg1
	Reg2
)

const (
	LoadWord = iota + 1
	StoreWord
	Add
	Sub
	BranchIfEqual
	AddImmediate = 7
	Halt         = 255
)

const (
	HaltJumpCode = iota + 20
	IncJumpCode
)

const (
	AddReg = iota
	AddImm
)

func main() {
	//mem := [255]int{0x01, 0x01, 0x10, 0x01,
	//	0x02, 0x12, 0x03, 0x01,
	//	0x02, 0x02, 0x01, 0x0e,
	//	0xff, 0x00,
	//	0x00, 0x00,
	//	0xa1, 0x14, 0x0c, 0x00}
	//
	//fmt.Println(&mem[1])
	//fmt.Println(&mem[2])
	//fmt.Println(1 << 64)
	//hold := 1 << 64
	//fmt.Println(hold)
	fmt.Printf("%x", (2<<16-255)&0xffff)

}

type Cpu struct {
	registers [3]int //0 will be pc
}

func (thisCpu *Cpu) load(register int, datum int) {
	thisCpu.registers[register] = datum
}

func (thisCpu *Cpu) store(register int) int {
	return thisCpu.registers[register]
}

func (thisCpu *Cpu) accumAddRegs(accumulatorReg int, addendReg int) {
	thisCpu.registers[accumulatorReg] = (thisCpu.registers[accumulatorReg] + thisCpu.registers[addendReg]) & 0xffff
}

func (thisCpu *Cpu) addImmediate(accumulatorReg int, addend int) {
	thisCpu.registers[accumulatorReg] = (thisCpu.registers[accumulatorReg] + addend) & 0xffff
}

func (thisCpu *Cpu) accumSubtractRegs(accumulatorReg int, subtrahendReg int) {
	thisCpu.registers[accumulatorReg] = (thisCpu.registers[accumulatorReg] - thisCpu.registers[subtrahendReg]) & 0xffff
}

func (thisCpu *Cpu) checkIfRegistersAreEqual(reg1 int, reg2 int) bool {
	return thisCpu.registers[reg1] == thisCpu.registers[reg2]
}

func (thisCpu *Cpu) FdeCycle(memory []int) {
	thisCpu.registers[PC] = 0
	for i := 0; i < 1000000; i++ {
		instruction := thisCpu.fetch(memory)
		execute := thisCpu.decode(instruction, memory)
		jumpCode := execute()

		if jumpCode == HaltJumpCode {
			break
		} else if jumpCode == IncJumpCode {
			thisCpu.registers[PC] += 3
		} else {
			thisCpu.registers[PC] = jumpCode
		}
	}
}

func (thisCpu *Cpu) fetch(memory []int) [3]int {
	instructionAddress := thisCpu.registers[PC]
	return [3]int{memory[instructionAddress],
		memory[instructionAddress+1],
		memory[instructionAddress+2]}
}

func (thisCpu *Cpu) decode(instruction [3]int, memory []int) func() int {
	instructionCode := instruction[0]
	param1 := instruction[1]
	param2 := instruction[2]
	var executeFunc func() int
	switch instructionCode & 0xf {
	case LoadWord:
		//instructionCode = "load"
		executeFunc = func() int {
			return loadWordOp(thisCpu, memory, param1, param2)
		}
	case StoreWord:
		executeFunc = func() int {
			return storeWordOp(thisCpu, memory, param1, param2)
		}
	case Add:
		executeFunc = func() int {
			return addOp(thisCpu, instructionCode, param1, param2)
		}
	case Sub:
		executeFunc = func() int {
			return subOp(thisCpu, param1, param2)
		}
	case BranchIfEqual:
		{
			executeFunc = func() int {
				return branchIfEqualOp(thisCpu, instructionCode, param1, param2)
			}
		}
	default:
		if instructionCode == Halt {
			executeFunc = func() int {
				return haltOp()
			}
			break
		}
		panic(errors.New("invalid instruction code"))
	}
	return executeFunc
}

func loadWordOp(cpu *Cpu, memory []int, register int, memAddress int) int {
	byte1, byte2 := memory[memAddress], memory[memAddress+1]
	cpu.load(register, byte1+byte2<<8)
	return IncJumpCode
}

func storeWordOp(cpu *Cpu, memory []int, register int, memAddress int) int {
	datum := cpu.store(register)
	byte1, byte2 := datum&0xff, datum>>8
	memory[memAddress], memory[memAddress+1] = byte1, byte2
	return IncJumpCode
}

func addOp(cpu *Cpu, instructionCode int, accumulatorReg int, addendParam int) int {
	addressMode := instructionCode >> 4
	if addressMode == AddReg {
		cpu.accumAddRegs(accumulatorReg, addendParam)
	} else if addressMode == AddImm {
		cpu.addImmediate(accumulatorReg, addendParam)
	} else {
		panic(errors.New("invalid add addressing mode"))
	}
	return IncJumpCode
}

func subOp(cpu *Cpu, accumulatorReg int, subtrahendReg int) int {
	cpu.accumSubtractRegs(accumulatorReg, subtrahendReg)
	return IncJumpCode
}

func branchIfEqualOp(cpu *Cpu, instructionCode int, reg1 int, reg2 int) int {
	isEqual := cpu.checkIfRegistersAreEqual(reg1, reg2)
	if isEqual {
		branchAddress := instructionCode >> 4
		return branchAddress
	}
	return IncJumpCode
}

func haltOp() int {
	return HaltJumpCode
}
