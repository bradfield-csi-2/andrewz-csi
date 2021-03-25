package vm

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

type vmCase struct{ x, y, out byte }
type vmTest struct {
	name  string
	asm   string
	cases []vmCase
}

var mainTests = []vmTest{
	// Do nothing, just halt
	{
		name: "Halt",
		asm: `
halt`,
		cases: []vmCase{{0, 0, 0}},
	},
	// Move a value from input to output
	{
		name: "LoadStore",
		asm: `
push 1
store 0
halt`,
		cases: []vmCase{
			{1, 0, 1},
			{255, 0, 255},
		},
	},
	// Add two unsigned integers together
	{
		name: "Add",
		asm: `
push 1
push 2
add
store 0
halt`,
		cases: []vmCase{
			{1, 2, 3},     // 1 + 2 = 3
			{254, 1, 255}, // support max int
			{255, 1, 0},   // correctly overflow
		},
	},
	{
		name: "Subtract",
		asm: `
push 2
push 1
sub
store 0
halt`,
		cases: []vmCase{
			{5, 3, 2},
			{0, 1, 255}, // correctly overflow backwards
		},
	},
}

var stretchGoalTests = []vmTest{
	// Support a basic jump, ie skipping ahead to a particular location
	{
		name: "Jump",
		asm: `
push 1
jump 14
store 0
halt`,
		cases: []vmCase{{42, 0, 0}},
	},
	// Support a "branch if equal to zero" with relative offsets
	{
		name: "Beqz",
		asm: `
push 1
push 2
beqz 3
pop
store 0
halt`,
		cases: []vmCase{
			{42, 0, 0},  // r2 is zero, so should branch over the store
			{42, 1, 42}, // r2 is nonzero, so should store back 42
		},
	},
	// Support adding immediate values
	{
		name: "Addi",
		asm: `
push 1
pushi 3
pushi 5
add
add
store 0
halt`,
		cases: []vmCase{
			{0, 0, 8},   // 0 + 3 + 5 = 8
			{20, 0, 28}, // 20 + 3 + 5 = 8
		},
	},
	// Calculate the sum of first n numbers (using subi to decrement loop index)
	{
		name: "Sum to n",
		asm: `
push 1
push 1
beqz 15
pop
pushi 1
push 1
sub
store 1
push 1
add
push 1
jump 12
pop
store 0
halt`,
		cases: []vmCase{
			{0, 0, 0},
			{1, 0, 1},
			{5, 0, 15},
			{10, 0, 55},
		},
	},
}

func TestCompute(t *testing.T) {
	for _, test := range mainTests {
		t.Run(test.name, func(t *testing.T) { testCompute(t, test) })
	}
	if os.Getenv("STRETCH") != "true" {
		println("Skipping stretch goal tests. Run `STRETCH=true go test` to include them.")
	} else {
		for _, test := range stretchGoalTests {
			t.Run(test.name, func(t *testing.T) { testCompute(t, test) })
		}
	}
}

// Given some assembly code and test cases, construct a program
// according to the required memory structure, and run in each
// case through the virtual machine
func testCompute(t *testing.T, test vmTest) {
	// assemble code and load into memory
	memory := make([]byte, 256)
	copy(memory[8:], assemble(test.asm))
	// for each case, set inputs and run vm
	for _, c := range test.cases {
		memory[1] = c.x
		memory[2] = c.y

		compute(memory)

		actual := memory[0]
		if actual != c.out {
			t.Fatalf("Expected f(%d, %d) to be %d, not %d", c.x, c.y, c.out, actual)
		}

		memory[1] = 0
		memory[2] = 0
	}
}

func reg(s string) (b byte) {
	return map[string]byte{
		"r1": 0x01,
		"r2": 0x02,
	}[s]
}

func mem(s string) (b byte) {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return byte(i)
}

func imm(s string) (b byte) {
	// for now, immediate values and memory addresses are both just ints
	return mem(s)
}

// Assemble the given assembly code to machine code
func assemble(asm string) []byte {
	mc := []byte{}
	asm = strings.TrimSpace(asm)
	for _, line := range strings.Split(asm, "\n") {
		parts := strings.Split(strings.TrimSpace(line), " ")
		switch parts[0] {
		case "push":
			mc = append(mc, []byte{0x01, imm(parts[1])}...)
		case "store":
			mc = append(mc, []byte{0x02, imm(parts[1])}...)
		case "add":
			mc = append(mc, []byte{0x03}...)
		case "sub":
			mc = append(mc, []byte{0x04}...)
		case "pushi":
			mc = append(mc, []byte{0x05, imm(parts[1])}...)
		case "pop":
			mc = append(mc, []byte{0x06}...)
		case "jump":
			mc = append(mc, []byte{0x07, imm(parts[1])}...)
		case "beqz":
			mc = append(mc, []byte{0x08, imm(parts[1])}...)
		case "halt":
			mc = append(mc, 0xff)
		default:
			panic("Invalid operation: " + parts[0])
		}
	}
	return mc
}
