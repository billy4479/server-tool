package lib

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ServerType uint8

const (
	Vanilla ServerType = iota
	Fabric
)

var serverStartTime *time.Time = nil

type Server struct {
	Name    string
	BaseDir string
	Version *VersionInfo
	Type    ServerType
	HasGit  bool
}

type GitProgress func() func(string)

func (s *Server) PrettyName() string {
	versionStr := s.Version.ID
	if s.Type == Fabric {
		versionStr += " on Fabric"
	}
	if s.HasGit {
		versionStr += " + Git"
	}
	return fmt.Sprintf("%s (%s)", s.Name, versionStr)
}

const (
	FabricJarName    = "fabric-server-launch.jar"
	VanillaJarName   = "server.jar"
	GitDirectoryName = ".git"

	minMemFlag = "-Xms%dM"
	maxMemFlag = "-Xmx%dM"
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

func ensureJavaPretty(s *Server, progress JavaDownloadProgress) (string, error) {
	L.Debug.Printf("\"%s\" requires Java %d\n", s.Name, s.Version.JavaVersion)
	javaExe, err := EnsureJavaIsInstalled(s.Version.JavaVersion, progress)
	if err != nil {
		return "", err
	}
	L.Ok.Printf("Java was found at \"%s\"\n", javaExe)
	return javaExe, nil
}

func runJar(s *Server, gui bool, javaProgress JavaDownloadProgress) error {
	javaExe, err := ensureJavaPretty(s, javaProgress)
	if err != nil {
		return err
	}

	args := []string{
		fmt.Sprintf(minMemFlag, C.Minecraft.Memory),
		fmt.Sprintf(maxMemFlag, C.Minecraft.Memory),
	}

	args = append(args, javaArgs...)

	if s.Type == Vanilla {
		args = append(args, VanillaJarName)
	} else if s.Type == Fabric {
		args = append(args, FabricJarName)
	} else {
		panic("HOW DID YOU DO THIS?")
	}

	if !gui {
		args = append(args, noGuiFlag)
	}

	java, err := filepath.Abs(javaExe)
	if err != nil {
		return err
	}

	return RunCmdPretty(
		s.BaseDir,
		java,
		args...,
	)
}

func (s *Server) Start(gui bool, javaProgress JavaDownloadProgress, gitProgress GitProgress) error {

	serverStartTime = new(time.Time)
	*serverStartTime = time.Now()

	if s.HasGit && C.Git.Enable {
		if err := PreFn(s.BaseDir, gitProgress); err != nil {
			return err
		}
	}

	err := runJar(s, gui, javaProgress)
	if err != nil {
		if err == ErrExitedAbnormally {
			L.Error.Println("The server terminated with an error. Git will not update. You should first go figure out what happened to the server then git-unfuck")
		}
		return err
	}

	if s.HasGit && C.Git.Enable {
		if err := PostFn(s.BaseDir, gitProgress); err != nil {
			return err
		}
	}

	serverStartTime = nil

	return nil
}

func FindServers(progress ManifestDownloadProgress) ([]Server, error) {
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
		isServer := false
		for _, entry := range entries {
			if !entry.IsDir() {
				if entry.Name() == VanillaJarName {
					possibleServerJar := filepath.Join(s.BaseDir, entry.Name())
					err = detectServerVersion(possibleServerJar, &s, progress)
					if err != nil {
						return nil, err
					}

					if s.Version == nil {
						// server.jar is not a Minecraft server
						break
					}
					isServer = true
				} else if entry.Name() == FabricJarName {
					s.Type = Fabric
				}
			} else {
				if entry.Name() == GitDirectoryName {
					s.HasGit = C.Git.Enable
				}
			}
		}

		s.Name = e.Name()
		if isServer {
			servers = append(servers, s)
		}
	}

	if len(servers) == 0 {
		L.Warn.Println("No servers were found")
	}

	return servers, nil
}

func detectServerVersion(serverJarPath string, s *Server, progress ManifestDownloadProgress) error {
	infos, err := GetVersionInfos(progress)
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

	L.Info.Printf("Downloading server jar for version %s\n", s.Version.ID)
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

	L.Ok.Println("Done!")

	if !C.Minecraft.NoEULA {
		eula, err := os.Create(filepath.Join(s.BaseDir, "eula.txt"))
		if err != nil {
			return err
		}
		defer eula.Close()
		_, err = eula.Write([]byte(eulaContent))
		if err != nil {
			return err
		}

		L.Ok.Println("Eula accepted")
	}

	L.Ok.Println("[+] Server created successfully!")

	return nil
}
