package main

import (
	_ "embed"
	"github.com/cli-ish/deezer-badge/internal/routes"
)

func main() {
	badgeServer := routes.BadgeServer{}
	badgeServer.Init()
	badgeServer.Start()
}
