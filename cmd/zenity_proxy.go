package cmd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/billy4479/server-tool/lib"
	"github.com/ncruces/zenity"
)

var defaultZenityOptions = []zenity.Option{
	zenity.Title(fmt.Sprintf("Server Tool %s", lib.Version)),
}

func setupIcon() error {
	iconPath := filepath.Join(lib.C.Application.CacheDir, "icon.png")

	writeIcon := func() error {
		lib.L.Debug.Printf("Writing icon at %s\n", iconPath)
		f, err := os.Create(iconPath)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.Write(icon)
		return err
	}

	f, err := os.Open(iconPath)
	if err != nil {
		err = writeIcon()
		if err != nil {
			return err
		}
	} else {
		defer f.Close()
		hasher := sha256.New()
		_, err := io.Copy(hasher, f)
		if err != nil {
			return err
		}
		hash := hasher.Sum(nil)

		if !bytes.Equal(hash, iconHash[:]) {
			f.Close()
			err := writeIcon()
			if err != nil {
				return err
			}
		}
	}

	defaultZenityOptions = append(defaultZenityOptions, zenity.WindowIcon(iconPath))
	return nil
}

func zenityList(text string, items []string, options ...zenity.Option) (string, error) {
	res, err := zenity.List(text, items, append(defaultZenityOptions, options...)...)

	lib.L.Debug.Printf(
		`zenity (list): text:"%s" items:%q result:"%s" error:"%s"`+"\n",
		text, items, res, err,
	)

	return res, err
}

func zenityEntry(text string, options ...zenity.Option) (string, error) {
	res, err := zenity.Entry(text, append(defaultZenityOptions, options...)...)

	lib.L.Debug.Printf(
		`zenity (entry): text:"%s" result:"%s" error:"%s"`+"\n",
		text, res, err,
	)

	return res, err
}

func zenityQuestion(text string, options ...zenity.Option) error {
	err := zenity.Question(text, append(defaultZenityOptions, options...)...)

	lib.L.Debug.Printf(
		`zenity (question): text:"%s" error:"%s"`+"\n",
		text, err,
	)

	return err
}

func zenityProgress(options ...zenity.Option) (zenity.ProgressDialog, error) {
	dialog, err := zenity.Progress(append(defaultZenityOptions, options...)...)

	lib.L.Debug.Printf(
		`zenity (dialog): error:"%s"`+"\n",
		err,
	)

	return dialog, err
}

func zenityInfo(text string, options ...zenity.Option) error {
	err := zenity.Info(text, append(defaultZenityOptions, options...)...)

	lib.L.Debug.Printf(
		`zenity (info): text:"%s" error:"%s"`+"\n",
		text, err,
	)

	return err
}

func zenityError(text string, options ...zenity.Option) error {
	err := zenity.Error(text, append(defaultZenityOptions, options...)...)

	lib.L.Debug.Printf(
		`zenity (error): text:"%s" error:"%s"`+"\n",
		text, err,
	)

	return err
}
