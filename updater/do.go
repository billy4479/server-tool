package updater

import (
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/billy4479/server-tool/logger"
)

func do(URL string) func() error {
	return func() error {
		logger.L.Info.Printf("[+] Downloading %s\n", URL)
		new, err := http.Get(URL)
		if err != nil {
			return err
		}
		defer new.Body.Close()

		tmp, err := os.CreateTemp("", "")
		if err != nil {
			return err
		}

		_, err = io.Copy(tmp, new.Body)
		if err != nil {
			tmp.Close()
			return err
		}

		err = tmp.Chmod(0700)
		tmp.Close()
		if err != nil {
			return err
		}

		logger.L.Ok.Println("[+] Done. Updating now")

		exe, err := os.Executable()
		if err != nil {
			return err
		}
		cmd := exec.Command(tmp.Name(), "replace", exe)
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		return cmd.Start()
	}
}
