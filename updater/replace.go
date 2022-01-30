package updater

import (
	"io"
	"os"
)

func AmITheUpdate(args []string) (bool, error) {
	if len(args) == 3 {
		if args[1] == "replace" {
			exe, err := os.Executable()
			if err != nil {
				return true, err
			}

			me, err := os.Open(exe)
			if err != nil {
				return true, err
			}
			defer me.Close()

			old, err := os.Create(args[2])
			if err != nil {
				return true, err
			}
			defer old.Close()

			_, err = io.Copy(old, me)
			if err != nil {
				return true, err
			}

			// config.C.Application.WorkingDir = filepath.Dir(args[2])
			return true, nil
		}
	}

	return false, nil
}
