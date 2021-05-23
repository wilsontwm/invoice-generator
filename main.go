package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	must(rootCmd.Execute())
}

func must(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "invoice-generator",
	Short:   "This is a tool that generates invoice on the fly",
	Version: "0.0.1",
}
