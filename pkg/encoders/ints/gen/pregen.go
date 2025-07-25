// Package main is a generator for the base10000 (4 digit) encoding of the ints
// library.
package main

import (
	"fmt"
	"orly.dev/pkg/utils/chk"
	"os"
)

func main() {
	fh, err := os.Create("pkg/ints/base10k.txt")
	if chk.E(err) {
		panic(err)
	}
	for i := range 10000 {
		fmt.Fprintf(fh, "%04d", i)
	}
}
