//go:build windows

package lib

import (
	"os/exec"
	"syscall"
)

func addSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
