package server

import (
	"fmt"
	"path/filepath"

	"github.com/billy4479/server-tool/config"
	"github.com/billy4479/server-tool/git"
	"github.com/billy4479/server-tool/java"
	"github.com/billy4479/server-tool/logger"
	"github.com/billy4479/server-tool/tui"
	"github.com/billy4479/server-tool/utils"
)

const (
	minMemFlag = "-Xms%d%s"
	maxMemFlag = "-Xmx%d%s"
	noGuiFlag  = "nogui"
)

var (
	// https://www.spigotmc.org/threads/guide-optimizing-spigot-remove-lag-fix-tps-improve-performance.21726/page-10#post-1055873
	javaArgs = []string{
		"-XX:+UseG1GC",
		"-XX:+UnlockExperimentalVMOptions",
		"-XX:MaxGCPauseMillis=50",
		"-XX:+DisableExplicitGC",
		"-XX:TargetSurvivorRatio=90",
		"-XX:G1NewSizePercent=50",
		"-XX:G1MaxNewSizePercent=80",
		"-XX:InitiatingHeapOccupancyPercent=10",
		"-XX:G1MixedGCLiveThresholdPercent=50",
		// "-XX:+AggressiveOpts",
		"-jar",
	}
)

func ensureJavaPretty(s *Server) (string, error) {
	if !config.C.Application.Quiet {
		logger.L.Info.Printf("[+] \"%s\" requires Java %d\n", s.Name, s.Version.JavaVersion)
	}
	javaExe, err := java.EnsureJavaIsInstalled(s.Version.JavaVersion)
	if err != nil {
		return "", err
	}
	if !config.C.Application.Quiet {
		logger.L.Ok.Printf("[+] Java was found at \"%s\"\n", javaExe)
	}
	return javaExe, nil
}

func runJar(s *Server) (bool, error) {
	var err error
	javaExe := config.C.Java.ExecutableOverride
	if javaExe == "" {
		javaExe, err = ensureJavaPretty(s)
		if err != nil {
			return false, err
		}
	}

	letter := func() string {
		if config.C.Java.Memory.Gigabytes {
			return "G"
		} else {
			return "M"
		}
	}
	args := []string{
		fmt.Sprintf(minMemFlag, config.C.Java.Memory.Amount, letter()),
		fmt.Sprintf(maxMemFlag, config.C.Java.Memory.Amount, letter()),
	}
	if !config.C.Java.Flags.OverrideDefault {
		args = append(args, javaArgs...)
	}
	args = append(args, config.C.Java.Flags.ExtraFlags...)

	if s.Type == Vanilla {
		args = append(args, VanillaJarName)
	} else if s.Type == Fabric {
		args = append(args, FabricJarName)
	} else {
		panic("HOW DID YOU DO THIS?")
	}

	if config.C.Minecraft.NoGUI {
		args = append(args, noGuiFlag)
	}

	java, err := filepath.Abs(javaExe)
	if err != nil {
		return false, err
	}

	return utils.RunCmdPretty(
		!config.C.Application.Quiet,
		false,
		s.BaseDir,
		config.C.Minecraft.Quiet,
		java,
		args...,
	)
}

func runScript(s *Server) error {
	if s.Version != nil {
		if _, err := ensureJavaPretty(s); err != nil {
			return err
		}
	} else if !config.C.Application.Quiet {
		logger.L.Warn.Println("[!] server.jar not found! Running the script blindly...")
	}

	if s.HasGit && !config.C.Application.Quiet {
		logger.L.Info.Printf("[+] \"%s\" supports Git but has a startup script: ignoring Git\n", s.Name)
	}

	cmdLine := filepath.Join(".", config.C.StartScript.Name)

	_, err := utils.RunCmdPretty(true, false, s.BaseDir, false, cmdLine)
	return err
}

func setStartFn(s *Server) {
	s.Start = func() error {
		if s.HasStartScript {
			if config.C.Application.Quiet {
				return runScript(s)
			}

			logger.L.Info.Printf("[?] \"%s\" has a startup script!\n", s.Name)
			opt, err := tui.MakeMenu(false,
				tui.Option{
					Description: "Run the script",
					Action: func() error {
						return runScript(s)
					},
				},
				tui.Option{
					Description: "No, stop",
					Action: func() error {
						logger.L.Ok.Println("[+] Stopping")
						return nil
					},
				},
			)
			if err != nil {
				return err
			}

			return opt.Action()
		}

		if s.HasGit && !config.C.Git.Disable {
			if err := git.PreFn(s.BaseDir); err != nil {
				return err
			}
		}

		success, err := runJar(s)
		if err != nil {
			return err
		}

		if s.HasGit && !config.C.Git.Disable {
			if !success {
				logger.L.Warn.Println("[?] The server terminated with an error. Update Git anyways?")
				opt, err := tui.MakeMenu(false,
					tui.Option{
						Description: "Yes, update git anyways",
						Action:      func() error { return git.PostFn(s.BaseDir) },
					},
					tui.Option{
						Description: "No, I'll do it manually (advanced but recommended)",
						Action:      func() error { return nil },
					},
				)
				if err != nil {
					return err
				}
				return opt.Action()
			}

			if err := git.PostFn(s.BaseDir); err != nil {
				return err
			}
		}

		return nil
	}
}
