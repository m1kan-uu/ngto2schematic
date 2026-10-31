//go:build !windows

package main

import (
	"fmt"
	"os"
)

// Windows 以外ではコマンドライン版:
//
//	ngto2schematic OUTDIR TARGET(1.7.10|1.12|rev) INPUT...
func main() {
	if len(os.Args) < 4 {
		fmt.Println("usage: ngto2schematic OUTDIR TARGET(1.7.10|1.12|rev) INPUT...")
		return
	}
	logf := func(s string) { fmt.Println(s) }
	if os.Args[2] == "rev" {
		RunAllRev(os.Args[3:], os.Args[1], "", logf, nil)
		return
	}
	RunAll(os.Args[3:], os.Args[1], os.Args[2], "", "", logf, nil)
}
