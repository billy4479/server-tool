package cmd

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/billy4479/server-tool/lib"
	"github.com/ncruces/zenity"
	"github.com/skratchdot/open-golang/open"
)

//go:embed icon.png
var icon []byte
var iconHash = sha256.Sum256(icon)

var defaultZenityOptions = []zenity.Option{
	zenity.Title("Server Tool"),
}

func moreOptions() error {
	options := []string{
		"Open server folder",
		"Edit config",
		"Open cache folder",
		"Wipe manifest cache",
		"Wipe java cache",
	}
	res, err := zenity.List("Advanced options", options, defaultZenityOptions...)
	lib.L.Debug.Printf("[+] zenity: \"%s\", err:%v\n", res, err)
	if err != nil || len(res) == 0 {
		return runGui()
	}

	switch res {
	case options[0]:
		return open.Start(lib.C.Application.WorkingDir)
	case options[1]:
		{
			configPath, _, err := lib.GetConfigPath()
			if err != nil {
				return err
			}
			return open.Start(configPath)
		}
	case options[2]:
		return open.Start(lib.C.Application.CacheDir)
	case options[3]:
		return os.RemoveAll(lib.ManifestPath())
	case options[4]:
		return os.RemoveAll(lib.JavaDir())
	}

	return nil
}

type manifestProgressGUI struct {
	total   int
	current int
	dialog  zenity.ProgressDialog
	cancel  func()

	sync.Mutex
}

func newManifestProgressGUI() *manifestProgressGUI {

	return &manifestProgressGUI{
		total:   0,
		current: 0,
		dialog:  nil, // This cannot be set without knowing the total
	}
}

func (p *manifestProgressGUI) SetTotal(total int) {
	p.Lock()
	defer p.Unlock()

	p.total = total
	dialog, err := zenity.Progress(append(defaultZenityOptions, zenity.MaxValue(total))...)

	if err != nil {
		panic(err)
	}

	p.dialog = dialog
}

func (p *manifestProgressGUI) Add(name string) {
	p.Lock()
	defer p.Unlock()

	p.current++
	if err := p.dialog.Value(p.current); err != nil {
		p.cancel()
		return
	}
	p.dialog.Text(name)
}

func (p *manifestProgressGUI) Done() {
	p.Lock()
	defer p.Unlock()

	p.dialog.Text("Done!")
	p.dialog.Complete()

	time.Sleep(100 * time.Millisecond)
	p.dialog.Close()
}

func (p *manifestProgressGUI) SetCancel(cancel func()) {
	p.cancel = cancel
}

func chooseServer() (*lib.Server, error) {
	servers, err := lib.FindServers(newManifestProgressGUI())
	if err != nil {
		return nil, err
	}

	createNewStr := "[create new]"
	serverNames := []string{}
	serverNames = append(serverNames, createNewStr)
	for _, v := range servers {
		serverNames = append(serverNames, v.PrettyName())
	}

	res, err := zenity.List("Select a server to start", serverNames, append(defaultZenityOptions, zenity.ExtraButton("More options"))...)
	lib.L.Debug.Printf("[+] zenity: \"%s\", err:%v\n", res, err)

	if err != nil {
		if err == zenity.ErrExtraButton {
			err = moreOptions()
			if err != nil {
				return nil, err
			}
			return chooseServer()
		}
		return nil, nil
	}
	if res == createNewStr {
		return createNew()
	}
	if res == "" {
		return chooseServer()
	}
	for _, v := range servers {
		if v.PrettyName() == res {
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
	name, err := zenity.Entry("Choose a name for the server", defaultZenityOptions...)
	lib.L.Debug.Printf("[+] zenity: \"%s\", err:%v\n", name, err)
	return name
}

func chooseVersion() (*lib.VersionInfo, error) {
	versions, err := lib.GetVersionInfosSorted(newManifestProgressGUI())
	if err != nil {
		return nil, err
	}
	versionNames := []string{}
	for _, v := range versions {
		if v.Type == lib.VersionTypeRelease {
			versionNames = append(versionNames, v.ID)
		}
	}
	res, err := zenity.List("Choose a minecraft version", versionNames, append(defaultZenityOptions, zenity.ExtraButton("More versions"))...)
	lib.L.Debug.Printf("[+] zenity: \"%s\", err:%v\n", res, err)
	if err == zenity.ErrExtraButton {
		ver, err := zenity.Entry("Select a Minecraft version", defaultZenityOptions...)
		lib.L.Debug.Printf("[+] zenity: \"%s\", err:%v\n", ver, err)
		if err != nil {
			return chooseVersion()
		}

		for _, v := range versions {
			if v.ID == ver {
				return &v, nil
			}
		}
	}

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

func serverOptions(s *lib.Server) error {
	res := zenity.Question(fmt.Sprintf("Server \"%s\" was selected", s.PrettyName()),
		append(defaultZenityOptions,
			zenity.OKLabel("Run"),
			zenity.CancelLabel("Cancel"),
			zenity.ExtraButton("More options"),
		)...)
	lib.L.Debug.Printf("[+] zenity: %v\n", res)

	switch res {
	case nil:
		return s.Start(true)
	case zenity.ErrCanceled:
		return res
	case zenity.ErrExtraButton:
		{
			options := []string{"Run", "Open folder", "Install Fabric"}
			res, err := zenity.List("More options", options, defaultZenityOptions...)
			lib.L.Debug.Printf("[+] zenity: \"%s\", err:%v\n", res, err)
			if err != nil || len(res) == 0 {
				return serverOptions(s)
			}
			switch res {
			case options[0]:
				return s.Start(true)
			case options[1]:
				return open.Start(s.BaseDir)
			case options[2]:
				_ = zenity.Info("Not yet implemented", defaultZenityOptions...)
				return nil
			}
		}
	}

	return nil
}

func checkUpdates() error {
	needUpdate, newVersionURL, err := lib.CheckUpdates()
	if err != nil {
		return err
	}

	if needUpdate {
		if lib.C.Application.AutoUpdate {
			err = zenity.Question("An update was found! Update now?",
				append(defaultZenityOptions, zenity.OKLabel("Update"), zenity.CancelLabel("I'll do it later"))...)

			if err == nil {
				if err = lib.DoUpdate(newVersionURL); err != nil {
					panic(err)
				}
				_ = zenity.Info("Restart the application to apply the update")
				os.Exit(0)
			}
		} else {
			lib.L.Info.Printf("Automatic updates are disabled, visit %s to download the update\n", newVersionURL)
		}
	}

	return nil
}

func runGui() error {

	if err := setupIcon(); err != nil {
		return err
	}

	if err := checkUpdates(); err != nil {
		lib.L.Warn.Printf("[!] An error has occurred while checking for updates: %v", err)
	}

	server, err := chooseServer()

	if err != nil {
		return err
	}
	if server == nil {
		return nil
	}

	if err = serverOptions(server); err == zenity.ErrCanceled {
		return runGui()
	}

	return err
}
