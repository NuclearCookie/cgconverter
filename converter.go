package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"io/ioutil"
)

func main() {
	var input string
	var output string
	ProcessArgs(&input, &output)
	input, output = ValidatePaths(&input, &output)

	b, err := ioutil.ReadFile(input)
	if err != nil {
		log.Fatal(err)
	}
	println(string(b))
}

func ProcessArgs(input, output *string) {
	flag.StringVar(input, "input", "", "The input file to convert")
	flag.StringVar(output, "output", "", "The converted output file")
	flag.Parse()
	if *input == "" {
		println("Please specify an input file by using the -input flag")
		os.Exit(0)
	}
}

func ValidatePaths(input, output *string) (string, string) {
	//convert symbolic links to the real path
	var newInput string
	var newOutput string
	var err error

	newInput, err = filepath.EvalSymlinks(*input)
	if err != nil {
		log.Fatal(err)
	}
	//check if the file exists
	file, err := os.Open(newInput)
	if err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}

	if *output == "" {
		newOutput = filepath.Join(filepath.Dir(newInput), "OnlineCGCode.txt")
		fmt.Printf("Generating output file at %s\n", newOutput)
	} else {
		newOutput, err = filepath.EvalSymlinks(*output)
		if err != nil {
			log.Fatal(err)
		}
	}
	return newInput, newOutput
}

func ConvertOfflineToOnline(input, output string) {

}
