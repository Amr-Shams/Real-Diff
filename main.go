package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
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
func isFunctionStart(line string) bool {
	// Regular expression to match function start
	functionStartRegex := regexp.MustCompile(`^\s*(?:\w+\s+)+\w+\s*\([^;]*\)\s*\{`)
	return functionStartRegex.MatchString(line)
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

	// Read the line
	reader := bufio.NewReader(f)

	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(line), nil
}

func main() {
	cFilePath := "../foo.c"

	functionNames, err := getFunctionNames(cFilePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	lineNumber := strings.Split(functionNames[0], " ")[1]
	//convert the string to int
	lineNumberInt := 0
	fmt.Sscanf(lineNumber, "%d", &lineNumberInt)
	fmt.Println(lineNumberInt)
	_, err = readLineFromFile(cFilePath, lineNumberInt)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

}
