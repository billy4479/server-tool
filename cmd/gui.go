package cmd

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/billy4479/server-tool/lib"
	"github.com/ncruces/zenity"
)

//go:embed icon.png
var icon []byte
var iconHash = sha256.Sum256(icon)

var defaultZenityOptions = []zenity.Option{
	zenity.Title("Server Tool"),
}

func chooseServer() (*lib.Server, error) {
	servers, err := lib.FindServers()
	if err != nil {
		return nil, err
	}

	createNewStr := "[create new]"
	serverNames := []string{}
	serverNames = append(serverNames, createNewStr)
	for _, v := range servers {
		serverNames = append(serverNames, v.Name)
	}

	res, err := zenity.List("Select a server to start", serverNames, defaultZenityOptions...)
	lib.L.Debug.Printf("\"%s\", err:%v\n", res, err)

	if err != nil {
		return nil, nil
	}
	if res == createNewStr {
		return createNew()
	}
	if res == "" {
		return chooseServer()
	}
	for _, v := range servers {
		if v.Name == res {
			return &v, nil
		}
	}
	panic("we should not be here")
}

func createNew() (*lib.Server, error) {
	name := chooseName()
	if len(name) == 0 {
		return chooseServer()
	}

	version, err := chooseVersion()
	if err != nil {
		return nil, err
	}
	if version == nil {
		return chooseServer()
	}

	server := &lib.Server{
		Name:    name,
		BaseDir: path.Join(lib.C.Application.WorkingDir, name),
		Version: version,
		Type:    lib.Vanilla,
		HasGit:  false,
	}

	return server, lib.CreateServer(server)
}

func chooseName() string {
	name, err := zenity.Entry("Choose a name for the server")
	lib.L.Debug.Printf("\"%s\", err:%v\n", name, err)
	return name
}

func chooseVersion() (*lib.VersionInfo, error) {
	versions, err := lib.GetVersionInfos()
	if err != nil {
		return nil, err
	}
	versionNames := []string{}
	for _, v := range versions {
		versionNames = append(versionNames, v.ID)
	}
	res, err := zenity.List("Choose a minecraft version", versionNames, zenity.DisallowEmpty())
	lib.L.Debug.Printf("\"%s\", err:%v\n", res, err)
	if err != nil {
		return nil, nil
	}

	for _, v := range versions {
		if res == v.ID {
			return &v, nil
		}
	}

	return chooseVersion()
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

func runGui() error {

	if err := setupIcon(); err != nil {
		return err
	}

	server, err := chooseServer()

	if err != nil {
		return err
	}
	if server == nil {
		return nil
	}

	return server.Start(true)
}
