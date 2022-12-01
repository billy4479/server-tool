//go:build windows

package cmd

import "github.com/TheTitanrain/w32"

func hideConsole() {
	w32.ShowWindow(w32.GetConsoleWindow(), w32.SW_HIDE)
}
