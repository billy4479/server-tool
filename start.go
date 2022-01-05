package main

import (
	"fmt"
	"path/filepath"
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
	if !config.Application.Quiet {
		Info.Printf("[+] \"%s\" requires Java %d\n", s.Name, s.Version.JavaVersion)
	}
	javaExe, err := ensureJavaIsInstalled(s.Version.JavaVersion)
	if err != nil {
		return "", err
	}
	if !config.Application.Quiet {
		Ok.Printf("[+] Java was found at \"%s\"\n", javaExe)
	}
	return javaExe, nil
}

func runJar(s *Server) (bool, error) {
	var err error
	javaExe := config.Java.ExecutableOverride
	if javaExe == "" {
		javaExe, err = ensureJavaPretty(s)
		if err != nil {
			return false, err
		}
	}

	letter := func() string {
		if config.Java.Memory.Gigabytes {
			return "G"
		} else {
			return "M"
		}
	}
	args := []string{
		fmt.Sprintf(minMemFlag, config.Java.Memory.Amount, letter()),
		fmt.Sprintf(maxMemFlag, config.Java.Memory.Amount, letter()),
	}
	if !config.Java.Flags.OverrideDefault {
		args = append(args, javaArgs...)
	}
	args = append(args, config.Java.Flags.ExtraFlags...)

	if s.Type == Vanilla {
		args = append(args, VanillaJarName)
	} else if s.Type == Fabric {
		args = append(args, FabricJarName)
	} else {
		panic("HOW DID YOU DO THIS?")
	}

	if config.Minecraft.NoGUI {
		args = append(args, noGuiFlag)
	}

	java, err := filepath.Abs(javaExe)
	if err != nil {
		return false, err
	}

	return runCmdPretty(!config.Application.Quiet, false, s.BaseDir, config.Minecraft.Quiet, java, args...)
}

func runScript(s *Server) error {
	if s.Version != nil {
		if _, err := ensureJavaPretty(s); err != nil {
			return err
		}
	} else if !config.Application.Quiet {
		Warn.Println("[!] server.jar not found! Running the script blindly...")
	}

	if s.HasGit && !config.Application.Quiet {
		Info.Printf("[+] \"%s\" supports Git but has a startup script: ignoring Git\n", s.Name)
	}

	cmdLine := filepath.Join(".", config.StartScript.Name)

	_, err := runCmdPretty(true, false, s.BaseDir, false, cmdLine)
	return err
}

func setStartFn(s *Server) {
	s.Start = func() error {
		if s.HasStartScript {
			if config.Application.Quiet {
				return runScript(s)
			}

			Info.Printf("[?] \"%s\" has a startup script!\n", s.Name)
			opt, err := makeMenu(false,
				Option{
					Description: "Run the script",
					Action: func() error {
						return runScript(s)
					},
				},
				Option{
					Description: "No, stop",
					Action: func() error {
						Ok.Println("[+] Stopping")
						return nil
					},
				},
			)
			if err != nil {
				return err
			}

			return opt.Action()
		}

		if s.HasGit && !config.Git.Disable {
			if err := gitPreFn(s); err != nil {
				return err
			}
		}

		success, err := runJar(s)
		if err != nil {
			return err
		}

		if s.HasGit && !config.Git.Disable {
			if !success {
				Warn.Println("[?] The server terminated with an error. Update Git anyways?")
				opt, err := makeMenu(false,
					Option{
						Description: "Yes, update git anyways",
						Action:      func() error { return gitPostFn(s) },
					},
					Option{
						Description: "No, I'll do it manually (advanced but recommended)",
						Action:      func() error { return nil },
					},
				)
				if err != nil {
					return err
				}
				return opt.Action()
			}

			if err := gitPostFn(s); err != nil {
				return err
			}
		}

		return nil
	}
}
