package cmd

import (
	"fmt"
	"github.com/sourcecodemage/yt-dl-utils/internal/usecase"
)

// CreateUserCmd defines the flags and positional arguments for `my-cli create-user`.
type RenameFilesCmd struct {
	SrcDir string `arg:"" help:"Path to the root directory" required:""`
	DstDir string `help:"Path to the destination directory for renamed files."`
	Force  bool   `help:"Force overwrite of existing NFO files" default:"false"`
}

// Run is automatically invoked by Kong when the create-nfo command is selected.
func (r *RenameFilesCmd) Run(globals *Globals) error {
	if globals.Verbose {
		fmt.Printf("Renaming files at %s. Force overwrite is set to %v\n", r.SrcDir, r.Force)
	}

	fmt.Println("RenameFiles command selected.")
	creator := &usecase.RenameFilesCreator{}
	if err := creator.Execute(r.SrcDir, r.DstDir, r.Force); err != nil {
		return fmt.Errorf("generate NFO files: %w", err)
	}
	fmt.Println("nfo command completed successfully.")
	return nil
}
