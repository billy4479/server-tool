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
	"github.com/dustin/go-humanize"
	"github.com/ncruces/zenity"
	"github.com/skratchdot/open-golang/open"
)

//go:embed icon.png
var icon []byte
var iconHash = sha256.Sum256(icon)

var defaultZenityOptions = []zenity.Option{
	zenity.Title(fmt.Sprintf("Server Tool %s", lib.Version)),
}

func moreOptions() error {
	options := []string{
		"Open server folder",
		"Open cache folder",
		"Open logs folder",
		"Edit config",
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
		return open.Start(lib.C.Application.CacheDir)
	case options[2]:
		return open.Start(lib.GetLogsPath())
	case options[3]:
		{
			configPath, _, err := lib.GetConfigPath()
			if err != nil {
				return err
			}
			return open.Start(configPath)
		}
	case options[4]:
		return os.RemoveAll(lib.ManifestPath())
	case options[5]:
		return os.RemoveAll(lib.JavaDir())
	}

	return nil
}

func unfuck(s *lib.Server) error {
	options := []string{"HELP! - Open documentation", "Manual save", "Reset from origin", "Remove lock"}

	res, err := zenity.List("Unfuck menu: BE CAREFUL!", options, defaultZenityOptions...)

	if err == zenity.ErrCanceled {
		return serverOptions(s)
	}

	switch res {
	case options[0]:
		err = open.Start("https://github.com/billy4479/server-tool/blob/master/Unfuck.md")
		break
	case options[1]:
		err = lib.UnfuckCommit(s.BaseDir)
		break
	case options[2]:
		err = lib.UnfuckReset(s.BaseDir)
		break
	case options[3]:
		err = lib.UnfuckRemoveLock(s.BaseDir)
		break
	}

	if err != nil {
		return err
	}

	return serverOptions(s)
}

type manifestProgressGUI struct {
	total   int
	current int
	dialog  zenity.ProgressDialog

	sync.Mutex
}

func (p *manifestProgressGUI) SetTotal(total int) {
	p.Lock()
	defer p.Unlock()

	p.total = total
	dialog, err := zenity.Progress(append(defaultZenityOptions, zenity.MaxValue(total), zenity.NoCancel())...)

	if err != nil {
		panic(err)
	}

	p.dialog = dialog
}

func (p *manifestProgressGUI) Add(name string) {
	p.Lock()
	defer p.Unlock()

	p.current++
	err := p.dialog.Value(p.current)
	if err != nil {
		return
	}
	err = p.dialog.Text(name)
	if err != nil {
		return
	}
}

func (p *manifestProgressGUI) Done() {
	p.Lock()
	defer p.Unlock()

	err := p.dialog.Text("Done!")
	if err != nil {
		return
	}
	err = p.dialog.Complete()
	if err != nil {
		return
	}

	time.Sleep(100 * time.Millisecond)
	p.dialog.Close()
}

func chooseServer() (*lib.Server, error) {
	servers, err := lib.FindServers(&manifestProgressGUI{})
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
	versions, err := lib.GetVersionInfosSorted(&manifestProgressGUI{})
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

type javaDownloadProgressGUI struct {
	total   string
	current uint64
	name    string
	dialog  zenity.ProgressDialog
}

func (p *javaDownloadProgressGUI) OnDownloadStart(size uint64, name string) {
	p.total = humanize.Bytes(size)
	p.name = name
	var err error
	p.dialog, err = zenity.Progress(append(defaultZenityOptions, zenity.MaxValue(int(size)), zenity.NoCancel())...)
	if err != nil {
		panic(err)
	}
}

func (p *javaDownloadProgressGUI) OnDownloadProgress(n int64) {
	p.current += uint64(n)

	err := p.dialog.Value(int(p.current))
	if err != nil {
		return
	}

	err = p.dialog.Text(fmt.Sprintf("Downloading %s (%s/%s)", p.name, humanize.Bytes(p.current), p.total))
	if err != nil {
		return
	}
}

func (p *javaDownloadProgressGUI) OnDownloadFinish() {
	if err := p.dialog.Complete(); err != nil {
		return
	}
	p.dialog.Close()
}

func (p *javaDownloadProgressGUI) OnExtractionStart(name string) {
	p.name = name
	var err error
	p.dialog, err = zenity.Progress(append(defaultZenityOptions, zenity.Pulsate(), zenity.NoCancel())...)
	if err != nil {
		panic(err)
	}
}

func (p *javaDownloadProgressGUI) OnExtractionProgress(name string) {
	if err := p.dialog.Text(fmt.Sprintf("Extracting %s", name)); err != nil {
		return
	}
}

func (p *javaDownloadProgressGUI) OnExtractionDone() {
	if err := p.dialog.Complete(); err != nil {
		return
	}
	p.dialog.Close()
}

func gitProgressGUI() func(string) {
	dialog, err := zenity.Progress(append(defaultZenityOptions, zenity.Pulsate(), zenity.NoCancel())...)

	// Unsupported: don't show any gui.
	if err != nil {
		return func(s string) {}
	}

	isComplete := false

	return func(s string) {
		if isComplete {
			lib.L.Warn.Println("[?] GitProgressGUI: calling function after dialog closed")
			return
		}

		if len(s) == 0 {
			dialog.Close()
			isComplete = true
		} else {
			err := dialog.Text(s)
			if err != nil {
				lib.L.Warn.Println("[?] GitProgressGUI:", err)
			}
		}
	}
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
		return s.Start(true, &javaDownloadProgressGUI{}, gitProgressGUI)
	case zenity.ErrCanceled:
		return res
	case zenity.ErrExtraButton:
		{
			options := []string{"Run", "Open folder", "Unfuck", "Install Fabric"}
			res, err := zenity.List("More options", options, defaultZenityOptions...)
			lib.L.Debug.Printf("[+] zenity: \"%s\", err:%v\n", res, err)
			if err != nil || len(res) == 0 {
				return serverOptions(s)
			}
			switch res {
			case options[0]:
				return s.Start(true, &javaDownloadProgressGUI{}, gitProgressGUI)
			case options[1]:
				return open.Start(s.BaseDir)
			case options[2]:
				return unfuck(s)
			case options[3]:
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

func runGui() (err error) {

	defer func() {
		if err != nil {
			zenity.Error(fmt.Sprintf("An error has occurred: %v", err), defaultZenityOptions...)
		}
	}()

	if err = setupIcon(); err != nil {
		return err
	}

	if err = checkUpdates(); err != nil {
		lib.L.Warn.Printf("[?] An error has occurred while checking for updates: %v", err)
		err = nil
	}

	err = runMainGui()
	return err
}

func runMainGui() error {
	server, err := chooseServer()

	if err != nil {
		return err
	}
	if server == nil {
		return nil
	}

	if err = serverOptions(server); err == zenity.ErrCanceled {
		return runMainGui()
	}

	return err
}
