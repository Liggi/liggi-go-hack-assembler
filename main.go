package main

import (
	"bufio"
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

	// Format output to be single string, separated by newlines
	outputString := ""
	for _, line := range output {
		outputString += line + "\n"
	}

	// Save to file
	outputFilename := fileName[:len(fileName)-3] + "hack"
	outputFile, _ := os.Create(outputFilename)
	outputFile.WriteString(outputString)
}
