//go:build !windows

package lib

import (
	"os/exec"
)

func addSysProcAttr(cmd *exec.Cmd) {}
