# Documentation for Code Comment Removal in Go

## Problem Statement

The main problem faced in the project was to remove comments from a C source code file. The comments could be either single-line comments starting with `//` or multi-line comments enclosed between `/*` and `*/`. The challenge was to handle these comments correctly, especially when they appear in the same line as the code.

Another problem was to handle the case where the function declaration and the opening brace `{` are on the same line. In the current implementation, the function declaration was being stored into the buffer and eventually written, which was not the desired behavior.

## Solution
The solution involved using a state machine to keep track of whether we're inside a comment or not. The state machine has three states:

- "0": We're not inside a comment.
- "2": We've encountered a '/' character and we're checking the next character to see if it's another '/' (for a single-line comment) or a '*' (for a multi-line comment).
- "mc": We're inside a multi-line comment.

The state machine starts in state "0". When it encounters a '/', it transitions to state "2". In state "2", if it encounters another '/', it transitions back to state "0" and ignores the rest of the line. If it encounters a '*', it transitions to state "mc" and ignores characters until it encounters a '*/' sequence, at which point it transitions back to state "0".

To handle the case where the function declaration and the opening brace `{` are on the same line, we modified the `processFunctions` function to check if the line contains an opening brace `{` when it detects the start of a function. If it does, it sets `insideFunction` to true and increments `depth`.

## Code

Here's the relevant part of the code:

```go
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
            if strings.Contains(line, "{") {
                insideFunction = true
                depth++
            }
        }
        if !startFunction {
            continue
        }
        // ... rest of your code ...
    }
}
```

### Conclusion
This solution effectively removes comments from a C source code file and handles the case where the function declaration and the opening brace `{` are on the same line. It uses a state machine to keep track of whether we're inside a comment or not, and it uses a buffer to store characters until it's sure they're not part of a comment.


## How to run the code

```bash
./main <input_file>
```
- expected the functions in the file 
- the functions names. 