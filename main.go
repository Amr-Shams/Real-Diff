package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stoicperlman/fls"
)

// Var that holds the info of the functionailty of the code
var removeAndExtract = &cobra.Command{
	Use:           "datedcoverage -m test_path -N newDate -O oldDate -f filterFile",
	Short:         "generate html report based on changes between 2 dates",
	RunE:          removeAndExtractFunctions,
	SilenceErrors: true,
	Long: strings.TrimSpace(`
The command datedcoverage is used to generate another version of the coverage.info file called dated_coverage.info by
keeping only the lines that where added or modified between 2 dates that the user should specify,
the command is a wrapper around  'gcov_gen_report'
		`),
}

// a global variable that holds the info of the command line arguments
var (
	// the path of the test directory
	testPath string
	// the new date
	newDate string
	// the old date
	oldDate string
	// the output file
	outputFile string
)

// Function represents a function in the C source code

type Function struct {
	Name string
	Body string
	Line int
}

// the main functionality the takes the arguments and the path of the C file and returns the functions in the file
func removeAndExtractFunctions(cmd *cobra.Command, args []string) error {
	// call the removeCommentsAndExtractFunctions function to get the functions in the C file from both dates using the testPath
	oldFile := fmt.Sprintf("%s/%s", testPath, oldDate)
	newFile := fmt.Sprintf("%s/%s", testPath, newDate)
	oldFunctions, err := removeCommentsAndExtractFunctions(oldFile)
	if err != nil {
		return err
	}
	newFunctions, err := removeCommentsAndExtractFunctions(newFile)
	if err != nil {
		return err
	}
	// fmt.Println("old functions")
	// for _, function := range oldFunctions {
	// 	fmt.Println(function.Name, function.Body, function.Line)
	// 	fmt.Println("------------------------------------------------------------------------------------------------------------------")
	// }
	// fmt.Println("new functions")
	// for _, function := range newFunctions {
	// 	fmt.Println(function.Name, function.Body, function.Line)
	// 	fmt.Println("------------------------------------------------------------------------------------------------------------------")
	// }
	// call the getChangedFunctions function to get the functions that are changed between the 2 dates
	changedFunctions, DeletedFunctions, AddedFunctions := getChangedFunctions(oldFunctions, newFunctions)
	// create a new file to write the functions that are changed between the 2 dates
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()
	// loop over the changed functions
	f.Write([]byte("Changed Functions\n"))
	for _, function := range changedFunctions {
		// write the function to the output file
		_, err := f.WriteString(fmt.Sprintf("%s %d\n", function.Name, function.Line))
		if err != nil {
			return err
		}
	}
	// loop over the Deleted functions
	f.Write([]byte("Deleted Functions\n"))
	for _, function := range DeletedFunctions {
		// write the function to the output file
		_, err := f.WriteString(fmt.Sprintf("%s %d\n", function.Name, function.Line))
		if err != nil {
			return err
		}
	}
	// loop over the Added functions
	f.Write([]byte("Added Functions\n"))
	for _, function := range AddedFunctions {
		// write the function to the output file
		_, err := f.WriteString(fmt.Sprintf("%s %d\n", function.Name, function.Line))
		if err != nil {
			return err
		}
	}
	return nil

}

// init function that takes the arguments from the command line
func init() {
	removeAndExtract.Flags().StringVarP(&testPath, "testPath", "m", "", "the path of the test directory")
	removeAndExtract.Flags().StringVarP(&newDate, "newDate", "N", "", "the new date")
	removeAndExtract.Flags().StringVarP(&oldDate, "oldDate", "O", "", "the old date")
	removeAndExtract.Flags().StringVarP(&outputFile, "outputFile", "f", "", "the output file")
	// mark some flags as required not to be empty
	removeAndExtract.MarkFlagRequired("testPath")
	removeAndExtract.MarkFlagRequired("newDate")
	removeAndExtract.MarkFlagRequired("oldDate")
	removeAndExtract.MarkFlagRequired("outputFile")

}

// Main functioon that takes the path of the C file as an argument
func main() {
	if err := removeAndExtract.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func removeCommentsAndExtractFunctions(filePath string) ([]Function, error) {
	f, err := fls.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.Seek(0, io.SeekStart)
	if err != nil && err != io.EOF {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	functions, err := getFunctions(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	lineNumbers := extractLineNumbers(functions)
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Line < functions[j].Line
	})

	processFunctions(scanner, functions, lineNumbers)
	return functions, nil
}

func getFunctions(cFilePath string) ([]Function, error) {
	// try to execute the ctags.sh script by the sudo command
	cmd1 := exec.Command("sudo", "./scripts/ctags.sh", cFilePath)
	err := cmd1.Run()
	if err != nil {
		return nil, fmt.Errorf("error running ctags: %v", err)
	}
	cmd2 := exec.Command("ctags", "-x", "-n", "--c-kinds=f", "--_xformat=%N %S %n", cFilePath)
	output, err := cmd2.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running ctags: %v", err)
	}

	functions := []Function{}
	for _, line := range strings.Split(string(output), "\n") {
		lineList := strings.Fields(line)
		if len(lineList) > 0 {
			lineNumber, err := strconv.Atoi(lineList[len(lineList)-1])
			if err != nil {
				return nil, fmt.Errorf("error parsing line number: %v", err)
			}
			functionSignature := strings.Join(lineList[:len(lineList)-1], " ")
			functions = append(functions, Function{Name: functionSignature, Line: lineNumber})
		}
	}

	return functions, nil
}

// func extractFunctionSignature(lineList []string) string {
// 	functionSignature := strings.Join(lineList[4:], " ")
// 	re := regexp.MustCompile(`^[\w\s]*\(\)`)
// 	matches := re.FindStringSubmatch(functionSignature)
// 	if len(matches) > 0 {
// 		functionSignature = matches[0]
// 	}
// 	return functionSignature
// }

func extractLineNumbers(functions []Function) []int {
	lineNumbers := []int{}
	for _, function := range functions {
		lineNumbers = append(lineNumbers, function.Line)
	}
	return lineNumbers
}

func processFunctions(scanner *bufio.Scanner, functions []Function, lineNumbers []int) {
	var insideFunction, startFunction bool
	var functionBody strings.Builder
	depth := 0
	index := 0
	var state = "0"
	var skipMultiLine bool
	for linNum := 1; scanner.Scan(); linNum++ {
		var skipLine bool = false
		line := scanner.Text()
		if !startFunction && isStartOfFunction(linNum, lineNumbers) {
			startFunction = true
		}
		if !startFunction {
			continue
		}
		var lastChar rune
		var buffer bytes.Buffer
		for _, char := range line {

			switch state {
			case "0":
				switch char {
				case '/':
					state = "2"
					buffer.WriteRune(lastChar)
					lastChar = char
					continue
				default:
					state = "0"
					buffer.WriteRune(lastChar)
					lastChar = char

				}
			case "2": // slash
				switch char {
				case '/': // single line comment
					state = "0"
					skipLine = true
					buffer.Reset()
				case '*': // multi line comment
					state = "mc"
					skipMultiLine = true
					buffer.Reset()
				default:
					state = "0"
					buffer.WriteRune(lastChar)
					lastChar = char
				}
			case "3": // end of the multi line comment
				switch char {
				case '/':
					state = "0"
					skipMultiLine = false
				default:
					state = "mc"
				}
			case "mc": // slash
				switch char {
				case '*':
					state = "3"
				}
			}

			if skipLine || skipMultiLine {
				continue
			}

			if char == '{' {
				depth++
				if !insideFunction {
					buffer.Reset()
					insideFunction = true
				}
			}

			if insideFunction && !skipLine && !skipMultiLine {
				functionBody.WriteString(buffer.String())
				buffer.Reset()
			}

			if char == '}' {
				depth--
				if depth == 0 && insideFunction {
					insideFunction = false
					startFunction = false
					// the follwoing line is to remove the spaces and new lines from the function body not human readable
					functions[index].Body = cleanFunctionBody(functionBody.String())
					//functions[index].Body = functionBody.String()
					index++
					functionBody.Reset()
					break
				}
			}

		}
		//functionBody.WriteString("\n")

	}
}

func isStartOfFunction(linNum int, lineNumbers []int) bool {
	for _, lineNumber := range lineNumbers {
		if linNum == lineNumber {
			return true
		}
	}
	return false
}

func cleanFunctionBody(functionBody string) string {
	functionBody = strings.ReplaceAll(functionBody, "\n", "")
	functionBody = strings.ReplaceAll(functionBody, "\t", "")
	functionBody = strings.ReplaceAll(functionBody, " ", "")
	return functionBody
}
func getChangedFunctions(oldFunctions []Function, newFunctions []Function) ([]Function, []Function, []Function) {
	oldFunctionMap := make(map[string]Function)
	newFunctionMap := make(map[string]Function)

	for _, function := range oldFunctions {
		oldFunctionMap[function.Name] = function
	}

	for _, function := range newFunctions {
		newFunctionMap[function.Name] = function
	}

	var changedFunctions []Function
	var addedFunctions []Function
	var deletedFunctions []Function

	for name, newFunction := range newFunctionMap {
		oldFunction, ok := oldFunctionMap[name]
		if !ok {
			addedFunctions = append(addedFunctions, newFunction)
		} else if oldFunction.Body != newFunction.Body {
			changedFunctions = append(changedFunctions, newFunction)
		}
	}

	for name, oldFunction := range oldFunctionMap {
		if _, ok := newFunctionMap[name]; !ok {
			deletedFunctions = append(deletedFunctions, oldFunction)
		}
	}

	return changedFunctions, addedFunctions, deletedFunctions
}
