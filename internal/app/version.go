package app

import (
	"fmt"
	"runtime"
)

// RunVersion prints version information.
func RunVersion() {
	fmt.Printf("taboo %s\n", Version)
	fmt.Printf("  commit:     %s\n", Commit)
	fmt.Printf("  built:      %s\n", BuildTime)
	fmt.Printf("  go version: %s\n", runtime.Version())
	fmt.Printf("  platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
