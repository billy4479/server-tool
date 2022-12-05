package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/billy4479/server-tool/lib"
	"github.com/urfave/cli/v2"
)

type UIMode uint8

const (
	GUI UIMode = iota
	TUI
	CLI
)

type manifestProgressCLI struct{}

func (*manifestProgressCLI) SetTotal(int)     {}
func (*manifestProgressCLI) Add(string)       {}
func (*manifestProgressCLI) Done()            {}
func (*manifestProgressCLI) SetCancel(func()) {}

type javaDownloadProgressCLI struct{}

func (p *javaDownloadProgressCLI) OnDownloadStart(size uint64, name string) {}
func (p *javaDownloadProgressCLI) OnDownloadProgress(n int64)               {}
func (p *javaDownloadProgressCLI) OnDownloadFinish()                        {}
func (p *javaDownloadProgressCLI) OnExtractionStart(name string)            {}
func (p *javaDownloadProgressCLI) OnExtractionProgress(name string)         {}
func (p *javaDownloadProgressCLI) OnExtractionDone()                        {}

func runCli() error {
	app := cli.App{
		Name:    "Server Tool",
		Version: lib.Version,
		Action: func(ctx *cli.Context) error {
			hideConsole()
			return runGui()
		},
		Usage: "Run and manage your Minecraft servers. If no command is specified runs in GUI mode",
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Action: func(ctx *cli.Context) error {
					lib.L.Debug.Printf("server-tool %s\n", lib.Version)
					return nil
				},
			},
			{
				Name:  "tui",
				Usage: "Run in TUI mode",
				Action: func(ctx *cli.Context) error {
					defer func() {
						if runtime.GOOS == "windows" {
							fmt.Print("Press enter to continue...")
							fmt.Scanln()
						}
					}()
					return runTui()
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List available servers",
				Action: func(ctx *cli.Context) error {
					servers, err := lib.FindServers(&manifestProgressCLI{})
					if err != nil {
						return err
					}
					for _, server := range servers {
						fmt.Println(server.PrettyName())
					}
					return nil
				},
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Server name",
						Aliases:  []string{"n"},
						Required: true,
					},
				},
				Usage: "Run a server",
				Action: func(ctx *cli.Context) error {
					servers, err := lib.FindServers(&manifestProgressCLI{})
					if err != nil {
						return err
					}

					name := ctx.String("name")
					for _, s := range servers {
						if s.Name == name {
							return s.Start(false, &javaDownloadProgressCLI{})
						}
					}
					return fmt.Errorf("Server %s not found", name)
				},
			},
		},
	}

	return app.Run(os.Args)
}
