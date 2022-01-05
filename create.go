package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const eulaContent = "eula=true"

func createServer(s *Server) error {
	err := os.MkdirAll(s.BaseDir, 0755)
	if err != nil {
		return err
	}

	Info.Printf("[+] Downloading jar for version %s\n", s.Version.ID)
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

	if !config.Application.Quiet {
		Ok.Println("[+] Done!")
	}
	if !config.Minecraft.NoEULA {
		Info.Println("[+] Accepting the EULA...")

		eula, err := os.Create(filepath.Join(s.BaseDir, "eula.txt"))
		if err != nil {
			return err
		}
		defer eula.Close()
		_, err = eula.Write([]byte(eulaContent))
		if err != nil {
			return err
		}

		if !config.Application.Quiet {
			Ok.Println("[+] Done!")
		}
	}
	return nil
}
