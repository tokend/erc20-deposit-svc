package main

import (
	"os"

	"github.com/tokend/erc20-deposit-svc/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
