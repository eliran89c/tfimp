package main

import (
	"fmt"
	"os"

	"github.com/eliran89c/tfimp/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil && err.Error() != "" {
		fmt.Println(err)
		os.Exit(1)
	}
}
