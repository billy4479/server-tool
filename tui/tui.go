package tui

import (
	"bufio"
	"os"
	"strings"
)

var inputReader = bufio.NewReader(os.Stdin)

func readLine() (string, error) {
	input, err := inputReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.ReplaceAll(input, "\r", "")
	input = strings.ReplaceAll(input, "\n", "")
	return input, nil
}
