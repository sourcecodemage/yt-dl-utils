package main

import (
	"log"

	"github.com/alecthomas/kong"
	"github.com/sourcecodemage/yt-dl-utils/internal/cmd"
)

// main function orchestrates the command-line interface and dispatches commands to their respective handlers.
func main() {
	cli := cmd.CLI{}
	// parse the command line
	ctx := kong.Parse(&cli,
		kong.Configuration(kong.JSON),
		kong.Name("yt-dl-utils"),
		kong.Description("A CLI wrapper to provide extra functionality to yt-dlp."),
		kong.UsageOnError(),
	)

	// if command.ConfigFilePath != "" {
	// 	appConfig := config.AppConfig{}
	// 	if err := appConfig.New(command.ConfigFilePath); err != nil {
	// 		log.Fatalf("failed to load application configuration: %v", err)
	// 	}
	// }
	ctx.Bind(&cli)
	if err := ctx.Run(); err != nil {
		log.Fatal(err)
	}
}
