package main

import (
	"bufio"
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

func main() {
	cFilePath := "../client.c"
	start := time.Now()
	err := cleanFile(cFilePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	elapsed := time.Since(start)
	fmt.Println("Time elapsed: ", elapsed)
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
	fmt.Println(functions)
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

	for linNum := 1; scanner.Scan(); linNum++ {
		line := scanner.Text()
		if !startFunction && isStartOfFunction(linNum, lineNumbers) {
			startFunction = true
		}
		if !startFunction {
			continue
		}

		for _, char := range line {
			if char == '{' {
				depth++
				if !insideFunction {
					insideFunction = true
				}
			}
			if insideFunction {
				functionBody.WriteRune(char)
			}
			if char == '}' {
				depth--
				if depth == 0 && insideFunction {
					insideFunction = false
					startFunction = false
					functionBodyString := cleanFunctionBody(functionBody.String())
					functions[index].Body = functionBodyString
					index++
					functionBody.Reset()
					break
				}
			}
		}
		functionBody.WriteRune('\n')
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
	functionBody = string(removeCStyleComments([]byte(functionBody)))
	fmt.Println("Function before Cpp style comment removal: ", functionBody)
	functionBody = string(removeCppStyleComments([]byte(functionBody)))
	fmt.Println("Function after Cpp style comment removal: ", functionBody)
	functionBody = strings.ReplaceAll(functionBody, "\n", "")
	functionBody = strings.ReplaceAll(functionBody, "\t", "")
	functionBody = strings.ReplaceAll(functionBody, " ", "")
	return functionBody
}

func removeCStyleComments(content []byte) []byte {
	ccmt := regexp.MustCompile(`/\*([^*]|[\r\n]|(\*+([^*/]|[\r\n])))*\*+/`)
	return ccmt.ReplaceAll(content, []byte(""))
}

func removeCppStyleComments(content []byte) []byte {
	ccmt := regexp.MustCompile(`(?m)//.*$`)
	return ccmt.ReplaceAll(content, []byte(""))
}
