package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

type Parser struct {
	symbolTable *SymbolTable
}

type SymbolTable struct {
	symbols map[string]uint16
}

var computations = map[string]uint16{
	"0":   0b0101010,
	"1":   0b0111111,
	"-1":  0b0111010,
	"D":   0b0001100,
	"A":   0b0110000,
	"!D":  0b0001101,
	"!A":  0b0110001,
	"-D":  0b0001111,
	"-A":  0b0110011,
	"D+1": 0b0011111,
	"A+1": 0b0110111,
	"D-1": 0b0001110,
	"A-1": 0b0110010,
	"D+A": 0b0000010,
	"D-A": 0b0010011,
	"A-D": 0b0000111,
	"D&A": 0b0000000,
	"D|A": 0b0010101,
	"M":   0b1110000,
	"!M":  0b1110001,
	"-M":  0b1110011,
	"M+1": 0b1110111,
	"M-1": 0b1110010,
	"D+M": 0b1000010,
	"D-M": 0b1010011,
	"M-D": 0b1000111,
	"D&M": 0b1000000,
	"D|M": 0b1010101,
}
var destinations = map[string]uint16{
	"M":   0b001,
	"D":   0b010,
	"DM":  0b011,
	"A":   0b100,
	"AM":  0b101,
	"AD":  0b110,
	"ADM": 0b111,
}
var jumps = map[string]uint16{
	"JGT": 0b001,
	"JEQ": 0b010,
	"JGE": 0b011,
	"JLT": 0b100,
	"JNE": 0b101,
	"JLE": 0b110,
	"JMP": 0b111,
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("No file specified")
	}

	fileName := os.Args[1]

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	parser := NewParser()

	output, err := parser.Parse(scanner)
	if err != nil {
		log.Fatal(err)
	}

	// Save to file
	extension := path.Ext(fileName)
	outputFilename := strings.TrimSuffix(fileName, extension) + ".hack"
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		log.Fatal(err)
	}

	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()

	for _, binary := range output {
		writer.WriteString(fmt.Sprintf("%016b", binary) + "\n")
	}
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols: initSymbols(),
	}
}

var nextSymbol uint16 = 16

func (sym *SymbolTable) Add(key string, value ...uint16) {
	address := nextSymbol

	if len(value) > 0 {
		address = value[0]
	} else {
		nextSymbol++
	}

	sym.symbols[key] = address
}

func (sym *SymbolTable) Contains(key string) bool {
	_, ok := sym.symbols[key]
	return ok
}

func (sym *SymbolTable) Get(key string) uint16 {
	return sym.symbols[key]
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

func NewParser() *Parser {
	return &Parser{
		symbolTable: NewSymbolTable(),
	}
}

func (p *Parser) Parse(scanner *bufio.Scanner) ([]uint16, error) {
	lineCount := 0
	instructions := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		if strings.HasPrefix(line, "(") && strings.HasSuffix(line, ")") {
			label := strings.Trim(line, "()")

			if p.symbolTable.Contains(label) {
				return nil, fmt.Errorf("duplicate label (%s)", label)
			}

			p.symbolTable.Add(label, uint16(lineCount))

			continue
		}

		lineCount++
		instructions = append(instructions, line)
	}

	binaries := []uint16{}

	for _, instruction := range instructions {
		binary, err := p.translateInstruction(instruction)
		if err != nil {
			return nil, err
		}

		binaries = append(binaries, binary)
	}

	return binaries, nil
}

func (p *Parser) translateInstruction(instruction string) (uint16, error) {
	instruction = strings.Split(instruction, "//")[0]
	instruction = strings.TrimSpace(instruction)

	if strings.HasPrefix(instruction, "@") {
		return p.translateAInstruction(instruction)
	} else {
		return p.translateCInstruction(instruction)
	}
}

func (p *Parser) translateAInstruction(instruction string) (uint16, error) {
	var location uint16
	address := strings.TrimPrefix(instruction, "@")

	parseduint, err := strconv.ParseUint(address, 10, 16)
	if err == nil {
		location = uint16(parseduint)
	} else {
		if !p.symbolTable.Contains(address) {
			p.symbolTable.Add(address)
		}

		location = p.symbolTable.Get(address)
	}

	return location, nil
}

func (p *Parser) translateCInstruction(instruction string) (uint16, error) {
	var dest, jump string
	var compBin, destBin, jumpBin uint16
	var ok bool

	if strings.Contains(instruction, "=") {
		s := strings.Split(instruction, "=")
		dest, instruction = s[0], s[1]

		destBin, ok = destinations[dest]
		if !ok {
			return 0, fmt.Errorf("unknown destination %s", dest)
		}
	}

	if strings.Contains(instruction, ";") {
		s := strings.Split(instruction, ";")
		instruction, jump = s[0], s[1]

		jumpBin, ok = jumps[jump]
		if !ok {
			return 0, fmt.Errorf("unknown jump %s", jump)
		}
	}

	compBin, ok = computations[instruction]
	if !ok {
		return 0, fmt.Errorf("unknown computation %s", instruction)
	}

	return 0b111<<13 | compBin<<6 | destBin<<3 | jumpBin, nil
}
