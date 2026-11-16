package main

import (
	"log"

	"github.com/alecthomas/kong"
	"github.com/sourcecodemage/yt-dl-utils/internal/cmd"
	"github.com/sourcecodemage/yt-dl-utils/internal/config"
)

// main function orchestrates the command-line interface and dispatches commands to their respective handlers.
func main() {
	command := cmd.CLI{}
	// parse the command line
	ctx := kong.Parse(&command,
		kong.Name("yt-dl-utils"),
		kong.Description("A CLI wrapper to provide extra functionality to yt-dlp."),
		kong.UsageOnError(),
	)

	if command.ConfigFilePath != "" {
		appConfig := config.AppConfig{}
		if err := appConfig.New(command.ConfigFilePath); err != nil {
			log.Fatalf("failed to load application configuration: %v", err)
		}
	}

	if err := ctx.Run(&command.Globals); err != nil {
		log.Fatal(err)
	}
}
