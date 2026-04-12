package main

import (
	"fmt"
	"os"

	"git.derfenix.pro/derfenix/archiveopds/internal/interfaces/cli"
)

func main() {
	if err := cli.NewRoot().Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "archiveopds: %v\n", err)
		os.Exit(1)
	}
}
