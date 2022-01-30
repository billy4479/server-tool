package updater

import (
	"os"

	"github.com/billy4479/server-tool/logger"
	"github.com/billy4479/server-tool/tui"
)

func Update() error {
	needUpdate, newVersionURL, err := checkUpdates()
	if err != nil {
		return err
	}

	if needUpdate {
		logger.L.Info.Println("[?] A new version as been found!")
		option, err := tui.MakeMenu(false,
			tui.Option{
				Description: "Update now",
				Action:      do(newVersionURL),
			},
			tui.Option{
				Description: "Do that later",
				Action:      func() error { return nil },
			},
		)

		if err != nil {
			return err
		}

		err = option.Action()
		if err != nil {
			return err
		}

		// Done, just use the new one
		os.Exit(2)
	}

	return nil
}
