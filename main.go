package main

import (
	"embed"
	"os"

	"github.com/crmin/pb-init/internal/initcli"
)

//go:embed templates/*
var templateFS embed.FS

func main() {
	os.Exit(initcli.Run(os.Args[1:], initcli.Env{
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		Templates: templateFS,
	}))
}
