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

	"github.com/stoicperlman/fls"
)

type Function struct {
	Name string
	Body string
	Line int
}

// write a main func that takes in a file path
func main() {
	start := time.Now()
	filePath := os.Args[1]
	err := cleanFile(filePath)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Time taken:", time.Since(start))
}

func cleanFile(filePath string) error {
	f, err := fls.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Seek(0, io.SeekStart)
	if err != nil && err != io.EOF {
		return err
	}

	scanner := bufio.NewScanner(f)
	functions, err := getFunctions(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	lineNumbers := extractLineNumbers(functions)
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Line < functions[j].Line
	})

	processFunctions(scanner, functions, lineNumbers)
	for _, function := range functions {
		fmt.Println("Function name:", function.Name, "Line number:", function.Line, "Function body:", function.Body)
	}
	return nil
}

func getFunctions(cFilePath string) ([]Function, error) {
	cmd := exec.Command("ctags", "-x", "--c-kinds=f", cFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running ctags: %v", err)
	}

	functions := []Function{}
	for _, line := range strings.Split(string(output), "\n") {
		lineList := strings.Fields(line)
		if len(lineList) > 0 {
			lineNumber, err := strconv.Atoi(lineList[2])
			if err != nil {
				return nil, fmt.Errorf("error converting line number to integer: %v", err)
			}
			functionSignature := extractFunctionSignature(lineList)
			functions = append(functions, Function{Name: functionSignature, Line: lineNumber})
		}
	}

	return functions, nil
}

func extractFunctionSignature(lineList []string) string {
	functionSignature := strings.Join(lineList[4:], " ")
	re := regexp.MustCompile(`^[\w\s]*\(\)`)
	matches := re.FindStringSubmatch(functionSignature)
	if len(matches) > 0 {
		functionSignature = matches[0]
	}
	return functionSignature
}

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
					insideFunction = true
				}
				buffer.Reset()
				buffer.WriteRune(char)
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
					functionBody.WriteString(";}")
					// the follwoing line is to remove the spaces and new lines from the function body not human readable
					//functions[index].Body = cleanFunctionBody(functionBody.String())
					functions[index].Body = functionBody.String()
					index++
					functionBody.Reset()
					break
				}
			}

		}
		functionBody.WriteString("\n")

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
