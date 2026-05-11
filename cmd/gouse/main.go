package main

import (
	"os"

	"github.com/zewillyan007/gouse/internal/cli"
)

// version is set by the linker on release builds via:
//   go build -ldflags "-X main.version=v0.1.0" ./cmd/gouse
var version = "dev"

func main() {
	os.Exit(cli.Execute(version))
}
