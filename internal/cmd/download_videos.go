package cmd

import (
	"fmt"
	"github.com/sourcecodemage/yt-dl-utils/internal/usecase"
)

// CreateUserCmd defines the flags and positional arguments for `my-cli create-user`.
type DownloadVideosCmd struct {
	DstDir string `help:"Path to the destination directory for downloaded files." required:""`
}

// Run is automatically invoked by Kong when the create-nfo command is selected.
func (d *DownloadVideosCmd) Run(globals *Globals) error {
	if globals.Verbose {
		fmt.Printf("DownloadVideoFiles NFO files at %s.\n", d.DstDir)
	}

	fmt.Println("DownloadVideoFiles command selected.")
	creator := &usecase.DownloadVideosCreator{}
	if err := creator.Execute(d.DstDir); err != nil {
		return fmt.Errorf("Download Video files: %w", err)
	}
	fmt.Println("DownloadVideoFiles command completed successfully.")
	return nil
}
