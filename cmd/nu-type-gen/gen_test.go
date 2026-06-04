package main

import (
	"fmt"
	"os"
	"testing"
)

func TestGenerator(t *testing.T) {
	gen, err := NewGenerator([]string{
		"./test/foo.nu",
		"./test/bar.nu",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer gen.Close()
	for _, f := range gen.Files {
		t.Log("----", f.Path)
		fmt.Println()
		fmt.Println()
		f.Generate(os.Stdout)
		fmt.Println()
		fmt.Println()
	}
}
