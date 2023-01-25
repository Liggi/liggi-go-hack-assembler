package main

import (
	"bufio"
	"fmt"
	"os"

	"liggis-hack-assembler/Parser"
)

func main() {
	if len(os.Args) < 2 {
		panic("No file specified")
	}

	fileName := os.Args[1]

	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	parser := Parser.NewParser()

	output := parser.Parse(scanner)

	fmt.Println(output)
}
