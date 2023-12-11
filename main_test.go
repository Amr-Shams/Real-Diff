package main

import (
	"strings"
	"testing"
	"time"
)

func BenchmarkRemoveComments(b *testing.B) {

	input := strings.Repeat("int main() { /* This is a comment */ return 0; }\n", b.N)
	for i := 0; i < b.N; i++ {
		startTime := time.Now()
		removeComments(input)
		elabsied := time.Since(startTime)
		b.Log(elabsied)
	}
}
