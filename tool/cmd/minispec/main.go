// CRC: crc-CLI.md
package main

import (
	"os"

	"github.com/anthropics/minispec/internal/cli"
)

func main() {
	c := &cli.CLI{}
	os.Exit(c.Run(os.Args[1:]))
}
