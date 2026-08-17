package main

import (
	"os"

	"github.com/qu1queee/cartitas/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
