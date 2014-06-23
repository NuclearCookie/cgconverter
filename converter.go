package main

import (
	"fmt"
	"flag"
	"os"
	"path/filepath"
	"log"
)

func main() {
	var input string
	var output string
	//process command line arguments
	ProcessArgs(&input, &output)
	input, output = ValidatePaths(&input, &output)
	
	
	/*data := make([]byte, 100)
	count, err := file.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("read %d bytes: %q\n", count, data[:count])*/
	
	fmt.Printf("input file: %s\n", input)
	fmt.Printf("output file: %s\n", output)
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
	_, err = os.Open(newInput)
	if err != nil {
		log.Fatal(err)
	}

	if *output == "" {
		newOutput = filepath.Join(filepath.Dir(newInput), "OnlineCGCode.txt")
		fmt.Printf("Generating output file at %s\n", newOutput)
	} else {
		newOutput, err =filepath.EvalSymlinks(*output)
		if err != nil {
			log.Fatal(err)
		}
	}
	return newInput, newOutput
}
func ConvertOfflineToOnline(input, output string) {

}