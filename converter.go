package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/nuclearcookie/substringfinder"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var inputFilePath string
	var outputFilePath string
	ProcessArgs(&inputFilePath, &outputFilePath)
	inputFilePath, outputFilePath = ValidatePaths(&inputFilePath, &outputFilePath)

	buffer, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	outputText := ConvertOfflineToOnline(&buffer)
	//permission 0644
	Output(outputFilePath, outputText)

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

func Output(path, text string) {
	//First write to file, then call fmt on the file, then copy the filecontent!
	ioutil.WriteFile(path, []byte(text), 0644)
	BuildAndFormat(path)
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	err = clipboard.WriteAll(string(buffer))
	if err != nil {
		log.Fatal(err)
	}
	println("Output file generated at ", path, "! Copied online code to clipboard!")
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
	newInput = *input
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
		newOutput = filepath.Join(filepath.Dir(newInput), "OnlineCGCode.go")
	} else {
		newOutput = *output
		/*newOutput, err = filepath.EvalSymlinks(*output)
		if err != nil {
			log.Fatal(err)
		}*/
		if filepath.Ext(*output) != ".go" {
			println("Please make the output file a .go file to allow us to format and test it!")
		}
	}
	return newInput, newOutput
}

func ConvertOfflineToOnline(buffer *[]byte) string {
	fileData := string(*buffer)
	inputChannelName := GetInputChannelName(fileData)
	outputChannelName := GetOutputChannelName(fileData)
	fileData = RemoveImport(fileData)
	fileData = RemoveCGReaderMainFunction(fileData)
	fileData = ImportMissingPackages(fileData)
	fileData = ReplaceOutputCalls(fileData, outputChannelName)
	fileData = ReplaceInputCalls(fileData, inputChannelName)
	//println(fileData)
	return fileData
}

//******************************************
// GETTERS
//******************************************
func GetInputChannelName(data string) string {
	start, _ := substringfinder.FindFirstOfSubString(data, "<-chan string", true)
	return substringfinder.GetLastWord(data[:start])
}

func GetOutputChannelName(data string) string {
	//go fmt removes spaces between string and )
	start, _ := substringfinder.FindFirstOfSubString(data, "chan string)", true)
	return substringfinder.GetLastWord(data[:start])
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
		start = strings.LastIndex(imports[:start], "\n")
		imports = imports[:start] + imports[end+1:]
	}
	data = strings.Replace(data, originalImportsBlock, imports, 1)
	return data

}

func GetImportBlock(data string) (int, int) {
	start, end := substringfinder.FindFirstOfSubString(data, "import", true)
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
	start, end = substringfinder.FindFirstOfSubString(data, "cgreader.RunManualPrograms", true)
	if start == -1 {
		start, end = substringfinder.FindFirstOfSubString(data, "cgreader.RunManualProgram", true)
	}
	if start == -1 {
		start, end = substringfinder.FindFirstOfSubString(data, "cgreader.RunAndValidateManualPrograms", true)
	}
	if start == -1 {
		start, end = substringfinder.FindFirstOfSubString(data, "cgreader.RunAndValidateManualProgram", true)
	}
	if start == -1 {
		println("Unknown cgreader main function.. cannot remove it!")
		os.Exit(0)
	}
	//Isolate the function
	_, end = substringfinder.FindIndicesBetweenMatchingRunesWithStartingIndex(data, '(', ')', end+1, true)
	return start, end
}

func ImportMissingPackages(data string) string {
	//add log to the imports block if it's not there already
	data = AddPackage(data, "log")
	data = AddPackage(data, "fmt")
	data = AddPackage(data, "bufio")
	data = AddPackage(data, "os")
	return data
}

func AddPackage(fileContent, packageName string) string {
	start, end := GetImportBlock(fileContent)
	importsBlock := fileContent[start : end+1]

	packageName = "\"" + packageName + "\""
	start, end = substringfinder.FindFirstOfSubString(importsBlock, packageName, false)
	//block not found: add it here
	if start == -1 && end == -1 {
		//copy a string..
		originalImportsBlock := importsBlock
		originalImportsBlock += " "
		originalImportsBlock = originalImportsBlock[:len(originalImportsBlock)-1]

		start, end = substringfinder.FindIndicesBetweenRunes(importsBlock, '(', ')')
		importsBlock = importsBlock[:end] + packageName + "\n" + importsBlock[end:]
		fileContent = strings.Replace(fileContent, originalImportsBlock, importsBlock, 1)
	}
	return fileContent
}

//******************************************
// DEBUG OUTPUT REPLACE BLOCK
//******************************************
func ReplaceOutputCalls(data, outputChannelName string) string {
	data = strings.Replace(data, "cgreader.Traceln", "log.Println", -1)
	data = strings.Replace(data, "cgreader.Tracef", "log.Printf", -1)
	data = strings.Replace(data, "cgreader.Trace", "log.Print", -1)
	data = strings.Replace(data, "cgreader.Println", "println", -1)
	data = strings.Replace(data, "cgreader.Printf", "fmt.Printf", -1)
	data = strings.Replace(data, "cgreader.Print", "fmt.Print", -1)

	outputChannelName += " <- "
	data = strings.Replace(data, outputChannelName+"fmt.Sprintf(", "fmt.Printf(", -1)

	start, end := 0, 0
	for ; start != -1 && end != -1; start, end = substringfinder.FindFirstOfSubStringWithStartingIndex(data, outputChannelName, end, true) {
		endLine := strings.Index(data[end:], "\n")
		endLine += end
		subStart, subEnd := substringfinder.FindIndicesBetweenRunes(data[end:endLine], '"', '"')
		if subStart != -1 && subEnd != -1 {
			startUnimplBrackets, endUnimplBrackets := substringfinder.FindIndicesBetweenRunes(data[end:endLine], '(', ')')
			if startUnimplBrackets != -1 && endUnimplBrackets != -1 && startUnimplBrackets < subStart && endUnimplBrackets > subEnd {
				fmt.Printf("Probably Unimplemented output function found at: %s\n", data[end:endLine])
			}
			data = data[:start] + "fmt.Printf(" + data[end+subStart:end+subEnd+1] + ")" + data[end+subEnd+1:]
		}
	}
	return data
}

func ReplaceInputCalls(data string, inputChannelName string) string {
	originalString := "fmt.Sscanln(<-"
	originalString += inputChannelName
	originalString += ","
	data = strings.Replace(data, originalString, "fmt.Scanln(", -1)
	start, end := substringfinder.FindFirstOfSubString(data, inputChannelName, true)
	for start != -1 && end != -1 {
		if substringfinder.IsWholeWord(data, start, end) {
			//<-inputchannelname == read a whole line from input
			if start > 2 {
				//go fmt makes sure there are no spaces between <- and input
				if data[start-2:start] == "<-" {
					//take everything left of this operation and save it...
					newLineIndex := strings.LastIndex(data[:start-2], "\n") + 1
					variableName := data[newLineIndex : start-2]
					//check if we need to add a scanner in this scope
					left, right := substringfinder.FindIndicesOfSurroundingRunesOfSubString(data, newLineIndex, end+1, '{', '}')
					newScannerString := "scanner := bufio.NewScanner(os.Stdin)"
					left, right = substringfinder.FindFirstOfSubString(data[left+1:right], newScannerString, true)
					if left == -1 && right == -1 {
						newScannerString += "\n"
					} else {
						fmt.Printf("Left and right are: %d and %d", left, right)
						newScannerString = ""
					}
					//remove the old line
					data = data[:newLineIndex] + newScannerString + "scanner.Scan()\n" + variableName + "scanner.Text()" + data[end+1:]
				}
			}
		}
		start, end = substringfinder.FindFirstOfSubStringWithStartingIndex(data, inputChannelName, end+1, true)
	}
	return data
}

//******************************************
// COMMAND BLOCK
//******************************************
func BuildAndFormat(file string) {
	if filepath.Ext(file) == ".go" {
		cmd := exec.Command("go", "fmt", file)
		cmd.Stdin = strings.NewReader(file)
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		cmd = exec.Command("go", "build", file)
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		println(out.String())
		if err != nil {
			println("Error building the output file.. Sorry, there must still be a little error in the converter!")
		}
	}
}
