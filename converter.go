package main

import (
	"flag"
	"fmt"
	"github.com/nuclearcookie/substringfinder"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var input string
	var output string
	ProcessArgs(&input, &output)
	input, output = ValidatePaths(&input, &output)

	buffer, err := ioutil.ReadFile(input)
	if err != nil {
		log.Fatal(err)
	}
	outputBuffer := ConvertOfflineToOnline(&buffer)
	//permission 0644 
	ioutil.WriteFile(output, outputBuffer, 0644)
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

func ConvertOfflineToOnline(buffer *[]byte) []byte {
	fileData := string(*buffer)
	fileData = RemoveImport(fileData)
	fileData = RemoveCGReaderMainFunction(fileData)
	println(fileData)
	return *buffer
}

//******************************************
// IMPORT BLOCK
//******************************************
func RemoveImport(data string) string {
	start, end := GetImportBlock(data)
	//end + 1 to include the last found rune
	//note: reslicing does not copy over the data!
	imports := data[start : end+1]
	//make a copy by adding 1 char and taking a slice of all -1	
	originalImportsBlock := imports
	originalImportsBlock += " "
	originalImportsBlock = originalImportsBlock[:len(originalImportsBlock)-1]
	//remove the cgreader import
	start, end = substringfinder.FindIndicesBetweenRunesContaining(imports, '"', '"', "cgreader")
	if start != -1 && end != -1 {
		imports = imports[0:start] + imports[end+1:]
	}
	data = strings.Replace(data, originalImportsBlock, imports, 1)
	return data

}

func GetImportBlock(data string) (int, int) {
	start, end := substringfinder.FindFirstOfSubString(data, "import")
	if start != -1 && end != -1 {
		if strings.IndexRune(data[start:], '(') < strings.IndexRune(data[start:], '"') {
			//import structure surrounded by brackets
			_, end = substringfinder.FindIndicesBetweenRunesWithStartingIndex(data, '(', ')', end)
		} else {
			_, end = substringfinder.FindIndicesBetweenRunesWithStartingIndex(data, '"', '"', end)
		}
	}
	if start == -1 || end == -1 {
		println("Something went wrong while finding the import block. Terminating..")
		os.Exit(0)
	}
	return start, end
}

//******************************************
// MAIN FUNCTION REMOVAL BLOCK
//******************************************
func RemoveCGReaderMainFunction(data string) string {
	start, end := GetCGReaderMainFunction(data)
	if start != -1 && end != -1 {
		mainString := data[start : end+1]
		originalMainString := mainString
		originalMainString += " "
		originalMainString = originalMainString[:len(originalMainString)-1]
		start, end = substringfinder.FindIndicesBetweenMatchingRunes(mainString, '{', '}', true)
		if start != -1 && end != -1 {
			mainString = mainString[start+1 : end]
			data = strings.Replace(data, originalMainString, mainString, 1)
		}
	}
	return data
}

func GetCGReaderMainFunction(data string) (int, int) {
	start, end := -1, -1
	start, end = substringfinder.FindFirstOfSubString(data, "cgreader.RunManualPrograms")
	if start == -1 {
		start, end = substringfinder.FindFirstOfSubString(data, "cgreader.RunManualProgram")
	}
	if start == -1 {
		start, end = substringfinder.FindFirstOfSubString(data, "cgreader.RunAndValidateManualPrograms")
	}
	if start == -1 {
		start, end = substringfinder.FindFirstOfSubString(data, "cgreader.RunAndValidateManualProgram")
	}
	if start == -1 {
		println("Unknown cgreader main function.. cannot remove it!")
		os.Exit(0)
	}
	//Isolate the function
	_, end = substringfinder.FindIndicesBetweenMatchingRunesWithStartingIndex(data, '(', ')', end+1, true)
	return start, end
}
