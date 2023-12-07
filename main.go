package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/stoicperlman/fls"
)

func getFunctionNames(cFilePath string) ([]string, error) {
	// Run ctags command to generate tags for the C file
	cmd := exec.Command("ctags", "-x", "--c-kinds=f", cFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running ctags: %v", err)
	}

	// Parse the output to extract function names
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	functionNames := make([]string, 0)

	for scanner.Scan() {
		line := scanner.Text()
		// fmt.Println(line)
		// split the words in the line into a list of words
		words := strings.Fields(line)
		fmt.Println(words[0], words[2])
		// append as well the third word which is the function line number
		functionNames = append(functionNames, words[0]+" "+words[2])

	}

	return functionNames, nil

}
func removeComments(input string) string {
	// Regular expression to match C-style comments
	commentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/|//.*?$`)
	return commentRegex.ReplaceAllString(input, "")
}

func readLineFromFile(filePath string, lineNumber int) (string, error) {
	f, err := fls.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Seek to the beginning of the specified line
	_, err = f.Seek(0, io.SeekStart)
	if err != nil && err != io.EOF {
		return "", err
	}

	scanner := bufio.NewScanner(f)
	var insideFunction bool

	// creat a buffer to store the function body
	functionBody := strings.Builder{}
	depth := 0

	for linNum := 1; scanner.Scan(); linNum++ {
		line := scanner.Text()
		for _, char := range line {
			if char == '{' {
				depth++
				if !insideFunction && linNum == lineNumber {
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
					break
				}
			}
		}
	}
	// remove comments from the function body
	// function body before removing comments
	fmt.Println(functionBody.String())

	return removeComments(functionBody.String()), nil
}

func main() {
	cFilePath := "../foo.c"

	functionNames, err := getFunctionNames(cFilePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, LineNumber := range functionNames {
		// convert the LineNumber to an int
		linNum, err := strconv.Atoi(strings.Split(LineNumber, " ")[1])
		println(linNum)
		functionBody, err := readLineFromFile(cFilePath, linNum)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println(functionBody)
	}
}
