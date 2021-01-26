package version

import (
	"fmt"
	"runtime"
)

var (
	version   = "v1.0"
	gitCommit = ""
)

func GetVersion() string {
	formatString := fmt.Sprintf("%-14s %%s\n%-14s %%s\n%-14s %%s", "Version", "Git Version", "Go Version")
	return fmt.Sprintf(formatString, version, gitCommit, runtime.Version())
}
