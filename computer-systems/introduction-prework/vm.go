package vm

const (
	Load  = 0x01
	Store = 0x02
	Add   = 0x03
	Sub   = 0x04
	Halt  = 0xff
)

// Stretch goals
const (
	Addi = 0x05
	Subi = 0x06
	Jump = 0x07
	Beqz = 0x08
)

const (
	PC = iota
	Reg1
	Reg2
)

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

	registers := [3]byte{8, 0, 0} // PC, R1 and R2

	// Keep looping, like a physical computer's clock
	for {

		instructionAddress := registers[PC]

		op, parm1, parm2 := memory[instructionAddress], memory[instructionAddress+1], memory[instructionAddress+2]

		// decode and execute
		switch op {
		case Load:
			registers[parm1] = memory[parm2]
		case Store:
			memory[parm2] = registers[parm1]
		case Add:
			registers[parm1] += registers[parm2]
		case Sub:
			registers[parm1] -= registers[parm2]
		case Addi:
			registers[parm1] += parm2
		case Subi:
			registers[parm1] -= parm2
		case Jump:
			registers[PC] = parm1
			continue
		case Beqz:
			if registers[parm1] == 0 {
				registers[PC] += parm2
			}
		case Halt:
			return
		}
		//if next length 3 instruction extends past heap boundary, exit
		if registers[PC] >= 253 {
			return
		}
		registers[PC] = registers[PC] + 3
	}
}
