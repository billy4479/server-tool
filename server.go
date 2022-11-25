package servertool

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type ServerType uint8

const (
	Vanilla ServerType = iota
	Fabric
)

type Server struct {
	Name           string
	BaseDir        string
	Version        *VersionInfo
	Type           ServerType
	HasGit         bool
	HasStartScript bool
	Start          func() error
}

const (
	FabricJarName    = "fabric-server-launch.jar"
	VanillaJarName   = "server.jar"
	GitDirectoryName = ".git"
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
	if !C.Application.Quiet {
		L.Info.Printf("[+] \"%s\" requires Java %d\n", s.Name, s.Version.JavaVersion)
	}
	javaExe, err := EnsureJavaIsInstalled(s.Version.JavaVersion)
	if err != nil {
		return "", err
	}
	if !C.Application.Quiet {
		L.Ok.Printf("[+] Java was found at \"%s\"\n", javaExe)
	}
	return javaExe, nil
}

func runJar(s *Server) (bool, error) {
	var err error
	javaExe := C.Java.ExecutableOverride
	if javaExe == "" {
		javaExe, err = ensureJavaPretty(s)
		if err != nil {
			return false, err
		}
	}

	letter := func() string {
		if C.Java.Memory.Gigabytes {
			return "G"
		} else {
			return "M"
		}
	}
	args := []string{
		fmt.Sprintf(minMemFlag, C.Java.Memory.Amount, letter()),
		fmt.Sprintf(maxMemFlag, C.Java.Memory.Amount, letter()),
	}
	if !C.Java.Flags.OverrideDefault {
		args = append(args, javaArgs...)
	}
	args = append(args, C.Java.Flags.ExtraFlags...)

	if s.Type == Vanilla {
		args = append(args, VanillaJarName)
	} else if s.Type == Fabric {
		args = append(args, FabricJarName)
	} else {
		panic("HOW DID YOU DO THIS?")
	}

	if C.Minecraft.NoGUI {
		args = append(args, noGuiFlag)
	}

	java, err := filepath.Abs(javaExe)
	if err != nil {
		return false, err
	}

	return RunCmdPretty(
		!C.Application.Quiet,
		false,
		s.BaseDir,
		C.Minecraft.Quiet,
		java,
		args...,
	)
}

func runScript(s *Server) error {
	if s.Version != nil {
		if _, err := ensureJavaPretty(s); err != nil {
			return err
		}
	} else if !C.Application.Quiet {
		L.Warn.Println("[!] server.jar not found! Running the script blindly...")
	}

	if s.HasGit && !C.Application.Quiet {
		L.Info.Printf("[+] \"%s\" supports Git but has a startup script: ignoring Git\n", s.Name)
	}

	cmdLine := filepath.Join(".", C.StartScript.Name)

	_, err := RunCmdPretty(true, false, s.BaseDir, false, cmdLine)
	return err
}

func setStartFn(s *Server) {
	s.Start = func() error {
		if s.HasStartScript {
			if C.Application.Quiet {
				return runScript(s)
			}

			L.Info.Printf("[?] \"%s\" has a startup script!\n", s.Name)
			opt, err := MakeMenu(false,
				Option{
					Description: "Run the script",
					Action: func() error {
						return runScript(s)
					},
				},
				Option{
					Description: "No, stop",
					Action: func() error {
						L.Ok.Println("[+] Stopping")
						return nil
					},
				},
			)
			if err != nil {
				return err
			}

			return opt.Action()
		}

		if s.HasGit && !C.Git.Disable {
			if err := PreFn(s.BaseDir); err != nil {
				return err
			}
		}

		success, err := runJar(s)
		if err != nil {
			return err
		}

		if s.HasGit && !C.Git.Disable {
			if !success {
				L.Warn.Println("[?] The server terminated with an error. Update Git anyways?")
				opt, err := MakeMenu(false,
					Option{
						Description: "Yes, update git anyways",
						Action:      func() error { return PostFn(s.BaseDir) },
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

			if err := PostFn(s.BaseDir); err != nil {
				return err
			}
		}

		return nil
	}
}

func MakeServersMenuItem(servers []Server) []Option {
	result := []Option{}

	for _, s := range servers {
		desc := fmt.Sprintf("\"%s\" (", s.Name)
		if s.Version == nil {
			desc += "?? on ??"
		} else {
			desc += fmt.Sprintf("%s on ", s.Version.ID)
			switch s.Type {
			case Vanilla:
				desc += "Vanilla"
			case Fabric:
				desc += "Fabric"
			}
		}

		if s.HasGit {
			desc += " - Git"
		}
		if s.HasStartScript {
			desc += ", Start Script"
		}

		desc += ")"

		result = append(result, Option{
			Description: desc,
			Action:      s.Start,
		})
	}

	return result
}

func FindServers() ([]Server, error) {
	serverDirs, err := os.ReadDir(C.Application.WorkingDir)
	if err != nil {
		return nil, err
	}

	servers := []Server{}

	for _, e := range serverDirs {
		var s Server
		s.BaseDir = filepath.Join(C.Application.WorkingDir, e.Name())

		if !e.IsDir() {
			continue
		}

		entries, err := os.ReadDir(s.BaseDir)
		if err != nil {
			return nil, err
		}

		s.Type = Vanilla
		for _, entry := range entries {
			if !entry.IsDir() {
				if entry.Name() == VanillaJarName {
					possibleServerJar := filepath.Join(s.BaseDir, entry.Name())
					err = detectServerVersion(possibleServerJar, &s)
					if err != nil {
						return nil, err
					}

					if s.Version == nil {
						// server.jar is not a Minecraft server
						break
					}
				} else if entry.Name() == FabricJarName {
					s.Type = Fabric
				} else if entry.Name() == C.StartScript.Name {
					s.HasStartScript = !C.StartScript.Disable
				}
			} else {
				if entry.Name() == GitDirectoryName {
					s.HasGit = !C.Git.Disable
				}
			}
		}
		if s.Version == nil && !s.HasStartScript {
			continue
		}

		s.Name = e.Name()
		setStartFn(&s)

		servers = append(servers, s)
	}

	if len(servers) == 0 {
		return servers, errors.New("No server were found!")
	}

	return servers, nil
}

func detectServerVersion(serverJarPath string, s *Server) error {
	infos, err := GetVersionInfos()
	if err != nil {
		return err
	}

	jar, err := os.Open(serverJarPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer jar.Close()

	hasher := sha1.New()
	_, err = io.Copy(hasher, jar)
	if err != nil {
		return err
	}
	sha := hex.EncodeToString(hasher.Sum(nil))

	for _, v := range infos {
		if v.SHA == sha {
			s.Version = &v
			return nil
		}
	}

	return nil
}

const eulaContent = "eula=true"

func CreateServer(s *Server) error {
	err := os.MkdirAll(s.BaseDir, 0755)
	if err != nil {
		return err
	}

	L.Info.Printf("[+] Downloading jar for version %s\n", s.Version.ID)
	jar, err := os.Create(filepath.Join(s.BaseDir, "server.jar"))
	if err != nil {
		return err
	}
	defer jar.Close()

	res, err := http.Get(s.Version.JarURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = io.Copy(jar, res.Body)
	if err != nil {
		return err
	}

	if !C.Application.Quiet {
		L.Ok.Println("[+] Done!")
	}
	if !C.Minecraft.NoEULA {
		L.Info.Println("[+] Accepting the EULA...")

		eula, err := os.Create(filepath.Join(s.BaseDir, "eula.txt"))
		if err != nil {
			return err
		}
		defer eula.Close()
		_, err = eula.Write([]byte(eulaContent))
		if err != nil {
			return err
		}

		if !C.Application.Quiet {
			L.Ok.Println("[+] Done!")
		}
	}
	return nil
}
