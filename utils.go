package main

import (
	"fmt"
	"os"
)

func getWorkDir() string {
	baseDir := os.Getenv("BASE_DIR")
	if baseDir != "" {
		return baseDir
	}

	return "."
}

func makeServersMenuItem(servers []Server) []option {
	result := []option{}

	for _, s := range servers {
		desc := fmt.Sprintf("%s (", s.Name)
		if s.Version == nil {
			desc += "?? on ??"
		} else {
			desc += fmt.Sprintf("%s on ", s.Version.ID)
			switch s.Type {
			case Vanilla:
				desc += "Vanilla"
			case Fabric:
				desc += "Fabric"
			case Forge:
				desc += "Forge"
			case Paper:
				desc += "PaperMC"
			}
		}

		if s.HasGit {
			desc += " - Git"
		}
		if s.HasStartScript {
			desc += ", Start Script"
		}

		desc += ")"

		result = append(result, option{
			Description: desc,
			Action:      s.Start,
		})
	}

	return result
}
