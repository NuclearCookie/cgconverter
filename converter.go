package main

import (
	"fmt"
	"flag"
	"os"
	"path"
)

func main() {
	var input string
	var output string
	flag.StringVar(&input, "input", "", "The input file to convert")
	flag.StringVar(&output, "output", "", "The converted output file")
	flag.Parse()
	if input == "" {
		println("Please specify an input file by using the -input flag")
		os.Exit(0)
	}

	if output == "" {
		output = path.Join(path.Dir(input), "OnlineCGCode.txt")
		fmt.Printf("Generating output file at %s\n", output)
	}
	fmt.Printf("input file: %s\n", input)
	fmt.Printf("output file: %s\n", output)
}
func ConvertOfflineToOnline(input, output string) {

}