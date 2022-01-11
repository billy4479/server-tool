package server

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/billy4479/server-tool/config"
	"github.com/billy4479/server-tool/logger"
)

const eulaContent = "eula=true"

func CreateServer(s *Server) error {
	err := os.MkdirAll(s.BaseDir, 0755)
	if err != nil {
		return err
	}

	logger.L.Info.Printf("[+] Downloading jar for version %s\n", s.Version.ID)
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

	if !config.C.Application.Quiet {
		logger.L.Ok.Println("[+] Done!")
	}
	if !config.C.Minecraft.NoEULA {
		logger.L.Info.Println("[+] Accepting the EULA...")

		eula, err := os.Create(filepath.Join(s.BaseDir, "eula.txt"))
		if err != nil {
			return err
		}
		defer eula.Close()
		_, err = eula.Write([]byte(eulaContent))
		if err != nil {
			return err
		}

		if !config.C.Application.Quiet {
			logger.L.Ok.Println("[+] Done!")
		}
	}
	return nil
}
