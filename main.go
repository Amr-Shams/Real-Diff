package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stoicperlman/fls"
)

var removeAndExtract = &cobra.Command{
	Use:           "datedcoverage -m <module_path> -N <new_date> -O <old_date> -f <filter_file> -s <src_files> -p <product> -c<scan>",
	Short:         "Generates a dated coverage report based on changes between two dates",
	RunE:          removeAndExtractFunctions,
	SilenceErrors: true,
	Long: strings.TrimSpace(`
The 'datedcoverage' command generates a dated coverage report by keeping only the lines that were added or modified between two specified dates. 
It takes four arguments: 
- module_path: The path to the module to be analyzed
- new_date: The end date for the period to be analyzed
- old_date: The start date for the period to be analyzed
- filter_file: The file that contains the list of the functions already covered by the test cases
- src_files: The list of the source files to be analyzed
- product: The product name
- scan: A boolean to check if the user want to scan for test cases
The command generates a new version of the 'coverage.info' file called 'dated_coverage.info'. 
This command is a wrapper around the 'gcov_gen_report' functionality.
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
	// the list of the src files
	srcFiles string
	// the product name
	product string
	// boolean to check if the user want to scan for test cases
	scan bool
)

// Function represents a function in the C source code

type Function struct {
	Name               string
	fullyQualifiedName string
	Body               string
	Line               int
}

// the main functionality the takes the arguments and the path of the C file and returns the functions in the file
func removeAndExtractFunctions(cmd *cobra.Command, args []string) error {
	// check if srcFiles is not empty
	if srcFiles == "" {
		return fmt.Errorf("srcFiles cannot be empty")
	}
	// check if we can open the srcFiles file
	srcFilesFile, err := os.Open(srcFiles)
	if err != nil {
		return err
	}
	// close the file
	defer srcFilesFile.Close()
	// create a scanner to read the srcFiles file
	scanner := bufio.NewScanner(srcFilesFile)

	// srcfiles list as string separated by comma
	var srcFilesList string
	var AllFunctions string
	// loop over the src files
	if testPath != "" {
		outputFile = testPath + "/" + outputFile
	}

	f, err := os.Create(outputFile + ".txt")
	if err != nil {
		return err
	}
	defer f.Close()
	// create another outputfile called summary
	f1, err := os.Create(outputFile + "_summary.txt")
	if err != nil {
		return err
	}
	defer f1.Close()
	var listSrcFiles []string // 1 means it is changed/add/deleted 0 means it is not
	for scanner.Scan() {
		// get the src file name
		result := scanner.Text()

		oldFile := "/wv/cal_nightly_TOT/" + oldDate + ".calibreube." + getWeekDay(oldDate) + "/ic/lv/src/" + result // the src file path in mgc home

		newFile := "/wv/cal_nightly_TOT/" + newDate + ".calibreube." + getWeekDay(newDate) + "/ic/lv/src/" + result // the src file path in mgc home

		// fmt.Println("oldFile:", oldFile)
		// fmt.Println("newFile:", newFile)
		// oldFile := testPath + "/" + oldDate + "/" + result // the src file path in mgc home
		// newFile := testPath + "/" + newDate + "/" + result // the src file path in mgc home
		oldFunctions, _ := removeCommentsAndExtractFunctions(oldFile)
		newFunctions, _ := removeCommentsAndExtractFunctions(newFile)
		// fmt.Println(oldFunctions)
		// fmt.Println(newFunctions)
		// write the functions to the a new file
		// f, err := os.Create(outputFile + "functions_before_after")
		// if err != nil {
		// 	return err
		// }
		// defer f.Close()

		// f.Write([]byte("Old Functions\n"))
		// for _, function := range oldFunctions {
		// 	// write the function to the output file
		// 	_, err := f.WriteString(fmt.Sprintf("%s %d\n", function.Name, function.Line))
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		// f.Write([]byte("New Functions\n"))
		// for _, function := range newFunctions {
		// 	// write the function to the output file
		// 	_, err := f.WriteString(fmt.Sprintf("%s %d\n", function.Name, function.Line))
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		// call the getChangedFunctions function to get the functions that are changed between the 2 dates
		changedFunctions, AddedFunctions, DeletedFunctions := getChangedFunctions(oldFunctions, newFunctions)
		// loop over the changed functions
		// check if the file has changed/added/deleted functions
		if quit := len(DeletedFunctions) + len(AddedFunctions) + len(changedFunctions); quit == 0 {
			listSrcFiles = append(listSrcFiles, result)
			fmt.Println("No changes in file:", result)
			continue
		}
		f.Write([]byte(result + "\n"))
		for idx, function := range changedFunctions {
			if idx == 0 {
				f.Write([]byte("Changed Functions\n"))
			}
			// write the function to the output file
			_, err := f.WriteString(fmt.Sprintf("%s %d\n", function.Name, function.Line))
			if err != nil {
				return err
			}
		}

		for idx, function := range DeletedFunctions {
			if idx == 0 {
				f.Write([]byte("Deleted Functions\n"))
			}
			// write the function to the output file
			_, err := f.WriteString(fmt.Sprintf("%s %d\n", function.Name, function.Line))
			if err != nil {
				return err
			}
		}
		// loop over the Added functions
		for idx, function := range AddedFunctions {
			if idx == 0 {
				f.Write([]byte("Added Functions\n"))
			}
			// write the function to the output file
			_, err := f.WriteString(fmt.Sprintf("%s %d\n", function.Name, function.Line))
			if err != nil {
				return err
			}
		}
		fmt.Println("the file has changed/added/deleted functions:", result)
		srcFilesList += result + ","
		for _, function := range changedFunctions {
			AllFunctions += function.fullyQualifiedName + ","
		}
		for _, function := range AddedFunctions {
			AllFunctions += function.fullyQualifiedName + ","
		}
		for _, function := range DeletedFunctions {
			AllFunctions += function.fullyQualifiedName + ","
		}

		f.Write([]byte("/*******************/\n"))
	}
	// remove the last comma
	if len(srcFilesList) > 0 {
		srcFilesList = srcFilesList[:len(srcFilesList)-1]
	}
	if len(AllFunctions) > 0 {
		AllFunctions = AllFunctions[:len(AllFunctions)-1]
	}
	writeToFileStr(outputFile+"srcFiles.txt", srcFilesList)

	writeToFileStr(outputFile+"functions.txt", AllFunctions)

	writeToFileList(outputFile+"_summary.txt", listSrcFiles)

	if scan == false {
		fmt.Println("No need to scan for test cases")
		return nil
	}

	testCases := getTestCases(AllFunctions, srcFilesList)
	writeToFileList(outputFile+"testCases", testCases)

	return nil
}

// init function that takes the arguments from the command line
func init() {
	removeAndExtract.Flags().StringVarP(&testPath, "testPath", "m", "", "the path of the test directory")
	removeAndExtract.Flags().StringVarP(&newDate, "newDate", "N", "", "the new date")
	removeAndExtract.Flags().StringVarP(&oldDate, "oldDate", "O", "", "the old date")
	removeAndExtract.Flags().StringVarP(&outputFile, "outputFile", "f", "", "the output file")
	removeAndExtract.Flags().StringVarP(&srcFiles, "srcFiles", "s", "", "the list of the src files")
	removeAndExtract.Flags().StringVarP(&srcFiles, "product", "p", "", "the product name")
	removeAndExtract.Flags().BoolVarP(&scan, "scan", "c", false, "scan for test cases")

	// mark some flags as required not to be empty
	removeAndExtract.MarkFlagRequired("newDate")
	removeAndExtract.MarkFlagRequired("oldDate")
	removeAndExtract.MarkFlagRequired("outputFile")
	removeAndExtract.MarkFlagRequired("srcFiles")

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
	cmd := exec.Command("./ctags/ctags", "-n", "--kinds-C++=f", "--fields=+{typeref}", "-o", "-", cFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running ctags: %v", err)
	}

	functions := []Function{}
	for _, line := range strings.Split(string(output), "\n") {
		lineList := strings.Fields(line)
		if len(lineList) == 0 {
			continue
		}
		functionName := lineList[0]
		if functionName == "operator" {
			functionName += lineList[1]
		}
		if len(lineList) < 2 {
			continue
		}
		// check if the unc is lambda function
		re := regexp.MustCompile(`__anon\w*`)
		if re.MatchString(functionName) {
			continue
		}
		var lineNumber int
		for _, field := range lineList {
			if strings.HasSuffix(field, ";\"") {
				lineNumberString := strings.TrimSuffix(field, ";\"")
				lineNumber, err = strconv.Atoi(lineNumberString)
				if err != nil {
					return nil, fmt.Errorf("error parsing line number: %v", err)
				}
				break
			}
		}
		// now itterate over the line list to get the class:
		var className string
		for _, field := range lineList {
			if strings.HasPrefix(field, "class:") {
				className = strings.TrimPrefix(field, "class:")
				className += "::"
				break
			}
		}
		// check in the className a string with __anno* replace it with (anonymous namespace)
		className = re.ReplaceAllString(className, "(anonymous namespace)")
		functions = append(functions, Function{Line: lineNumber, fullyQualifiedName: className + functionName, Name: functionName})
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
	functionBody = strings.ReplaceAll(functionBody, "\r", "")
	functionBody = strings.ReplaceAll(functionBody, "\x00", "")
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
	// loop over the new functions
	for name, newFunction := range newFunctionMap {
		oldFunction, ok := oldFunctionMap[name]
		// if the function is not in the old functions then it is added
		if !ok {
			addedFunctions = append(addedFunctions, newFunction)
		} else if oldFunction.Body != newFunction.Body {
			// if the function is in the old functions and the body is not the same then it is changed
			changedFunctions = append(changedFunctions, newFunction)
		}
	}

	for name, oldFunction := range oldFunctionMap {
		// if the function is not in the new functions then it is deleted
		if _, ok := newFunctionMap[name]; !ok {
			deletedFunctions = append(deletedFunctions, oldFunction)
		}
	}

	return changedFunctions, addedFunctions, deletedFunctions
}

// function that takes a date and returns the weekday
func getWeekDay(date string) string {
	date_list := strings.Split(date, ".")
	date_list_int := [3]int{}
	for i, val := range date_list {
		intVar, _ := strconv.Atoi(val)
		date_list_int[i] = intVar
	}
	d := time.Date(date_list_int[0], time.Month(date_list_int[1]), date_list_int[2], 0, 0, 0, 0, time.Local)
	return strings.ToLower(d.Weekday().String())

}

//////////////////////////////
/*
*@TODO:
- add a function that takes the list of functions and returns thier test cases from the db.
- store the result in a set and return it
- then we can count how many functions are covered by the test cases
*/
func getTestCases(functions string, srcFiles string) []string {

	cmdArgs := []string{"gogcov", "search", "testcases", "--srcfiles-file", outputFile + "srcFiles.txt", "--functions-file", outputFile + "functions.txt"}
	if product != "" {
		cmdArgs = append(cmdArgs, "--products", product)
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
	}
	// split the output by new line
	testCases := strings.Split(string(output), "\n")
	if len(testCases) > 0 {
		// remove the last element which is empty
		testCases = testCases[:len(testCases)-1]
	}

	return testCases
}

//////////////////////////////
/*
* function to write to an output file
 */
func writeToFileList(outputFile string, testCases []string) {
	// create a new file to write the functions that are changed between the 2 dates
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer f.Close()
	// loop over the changed functions
	for _, testCase := range testCases {
		// write the function to the output file
		_, err := f.WriteString(fmt.Sprintf("%s\n", testCase))
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

//////////////////////////////
/*
* function to write a string to an output file
 */

func writeToFileStr(outputFile string, functions string) {
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer f.Close()
	functions = strings.ReplaceAll(functions, ",", "\n")
	_, err = f.WriteString(fmt.Sprintf("%s\n", functions))
	if err != nil {
		fmt.Println("Error:", err)
	}

}
