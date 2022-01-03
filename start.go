package main

import (
	"path/filepath"
)

var (
	// https://www.spigotmc.org/threads/guide-optimizing-spigot-remove-lag-fix-tps-improve-performance.21726/page-10#post-1055873
	javaArgs = []string{
		"-Xms6G",
		"-Xmx6G",
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
	Info.Printf("[+] \"%s\" requires Java %d\n", s.Name, s.Version.JavaVersion)
	javaExe, err := ensureJavaIsInstalled(s.Version.JavaVersion)
	if err != nil {
		return "", err
	}
	Ok.Printf("[+] Java was found at \"%s\"\n", javaExe)

	return javaExe, nil
}

func runDefaultJar(s *Server) (bool, error) {
	javaExe, err := ensureJavaPretty(s)
	if err != nil {
		return false, err
	}

	args := javaArgs
	if s.Type == Vanilla {
		args = append(args, VanillaJarName)
	} else if s.Type == Fabric {
		args = append(args, s.BaseDir, VanillaJarName)
	} else {
		panic("TODO")
	}

	java, err := filepath.Abs(javaExe)
	if err != nil {
		return false, err
	}

	return runCmdPretty(true, false, s.BaseDir, java, args...)
}

func setStartFn(s *Server) {
	s.Start = func() error {

		if s.HasStartScript {
			Info.Printf("[?] \"%s\" has a startup script!\n", s.Name)
			opt, err := makeMenu(false,
				Option{
					Description: "Run the script",
					Action: func() error {
						if s.Version != nil {
							if _, err := ensureJavaPretty(s); err != nil {
								return err
							}
						} else {
							Warn.Println("[!] server.jar not found! Running the script blindly...")
						}

						if s.HasGit {
							Info.Printf("[+] \"%s\" supports Git but has a startup script: ignoring Git\n", s.Name)
						}

						cmdLine := filepath.Join(".", StartScriptName)

						_, err := runCmdPretty(true, false, s.BaseDir, cmdLine)
						return err
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

		if s.HasGit {
			if err := gitPreFn(s); err != nil {
				return err
			}
		}

		success, err := runDefaultJar(s)
		if err != nil {
			return err
		}

		if s.HasGit {
			if !success {
				Warn.Println("[?] The server terminated with an error. Update Git anyways?")
				opt, err := makeMenu(false,
					Option{
						Description: "Yes, update git anyways",
						Action:      func() error { return gitPostFn(s) },
					},
					Option{
						Description: "No, I'll do it manually (ADVANCED)",
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
