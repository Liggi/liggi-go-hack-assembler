package Parser

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	symbolTable *SymbolTable
}

type SymbolTable struct {
	symbols map[string]uint16
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols: initSymbols(),
	}
}

var nextSymbol uint16 = 16
var computations = map[string]string{
	"0":   "0101010",
	"1":   "0111111",
	"-1":  "0111010",
	"D":   "0001100",
	"A":   "0110000",
	"!D":  "0001101",
	"!A":  "0110001",
	"-D":  "0001111",
	"-A":  "0110011",
	"D+1": "0011111",
	"A+1": "0110111",
	"D-1": "0001110",
	"A-1": "0110010",
	"D+A": "0000010",
	"D-A": "0010011",
	"A-D": "0000111",
	"D&A": "0000000",
	"D|A": "0010101",
	"M":   "1110000",
	"!M":  "1110001",
	"-M":  "1110011",
	"M+1": "1110111",
	"M-1": "1110010",
	"D+M": "1000010",
	"D-M": "1010011",
	"M-D": "1000111",
	"D&M": "1000000",
	"D|M": "1010101",
}
var destinations = map[string]string{
	"M":   "001",
	"D":   "010",
	"DM":  "011",
	"A":   "100",
	"AM":  "101",
	"AD":  "110",
	"ADM": "111",
}
var jumps = map[string]string{
	"JGT": "001",
	"JEQ": "010",
	"JGE": "011",
	"JLT": "100",
	"JNE": "101",
	"JLE": "110",
	"JMP": "111",
}

func (sym *SymbolTable) Add(key string, value ...uint16) {
	if len(value) == 0 {
		value[0] = nextSymbol
		nextSymbol++
	}

	sym.symbols[key] = value[0]
}

func (sym *SymbolTable) Contains(key string) bool {
	_, ok := sym.symbols[key]
	return ok
}

func (sym *SymbolTable) Get(key string) (uint16, error) {
	if !sym.Contains(key) {
		return 0, fmt.Errorf("symbol %s not found", key)
	}

	return sym.symbols[key], nil
}

func initSymbols() map[string]uint16 {
	symbols := map[string]uint16{
		"SP":     0,
		"LCL":    1,
		"ARG":    2,
		"THIS":   3,
		"THAT":   4,
		"SCREEN": 0x4000,
		"KBD":    0x6000,
	}

	// R0-R15
	for i := uint16(0); i <= 15; i++ {
		key := fmt.Sprintf("R%d", i)
		value := i
		symbols[key] = value
	}

	return symbols
}

var p *Parser

func NewParser() *Parser {
	p = &Parser{
		symbolTable: NewSymbolTable(),
	}

	return p
}

func (p *Parser) Parse(scanner *bufio.Scanner) map[string]string {
	instructions := map[string]string{}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		instruction := translateInstruction(line)

		instructions[line] = instruction
	}

	return instructions
}

func translateInstruction(instruction string) string {
	fmt.Println(instruction)
	if strings.HasPrefix(instruction, "(") {
		return "" // It's a label, we'll handle it later
	}

	if strings.HasPrefix(instruction, "@") {
		return translateAInstruction(instruction)
	} else {
		return translateCInstruction(instruction)
	}
}

func translateAInstruction(instruction string) string {
	var location uint16
	address := strings.TrimPrefix(instruction, "@")

	parseduint, err := strconv.ParseUint(address, 10, 16)
	if err == nil {
		location = uint16(parseduint)
	} else {
		return "" // It's a symbol, we'll handle it later
	}

	binary := fmt.Sprintf("%016b", location)

	return binary
}

func translateCInstruction(instruction string) string {
	var dest, jump string
	var destBin, compBin, jumpBin string
	var ok bool

	if strings.Contains(instruction, "=") {
		s := strings.Split(instruction, "=")
		dest, instruction = s[0], s[1]

		destBin, ok = destinations[dest]
		if !ok {
			destBin = "000"
		}
	}

	if strings.Contains(instruction, ";") {
		s := strings.Split(instruction, ";")
		instruction, jump = s[0], s[1]

		jumpBin, ok = jumps[jump]
		if !ok {
			jumpBin = "000"
		}
	}

	compBin, ok = computations[instruction]
	if !ok {
		panic(fmt.Sprintf("Unknown computation %s", instruction))
	}

	return fmt.Sprintf("111%s%s%s", compBin, destBin, jumpBin)
}
