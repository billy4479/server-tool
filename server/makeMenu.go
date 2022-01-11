package server

import (
	"fmt"

	"github.com/billy4479/server-tool/tui"
)

func MakeServersMenuItem(servers []Server) []tui.Option {
	result := []tui.Option{}

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

		result = append(result, tui.Option{
			Description: desc,
			Action:      s.Start,
		})
	}

	return result
}
