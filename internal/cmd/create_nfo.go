package cmd

import (
	"fmt"
	"github.com/sourcecodemage/yt-dl-utils/internal/usecase"
)

// CreateUserCmd defines the flags and positional arguments for `my-cli create-user`.
type CreateNFOCmd struct {
	// RootDir string `arg:"" help:"Path to the root directory" required:""`
	Force bool `help:"Force overwrite of existing NFO files" optional:"" default:"false"`
}

// Run is automatically invoked by Kong when the create-nfo command is selected.
func (c *CreateNFOCmd) Run(globals *Globals) error {
	if globals.Verbose {
		fmt.Printf("Creating NFO files at %s. Force overwrite is set to %v\n", c.RootDir, c.Force)
	}

	fmt.Println("NFO command selected.")
	creator := &usecase.NFOCreator{}
	if err := creator.Execute(c.RootDir, c.Force); err != nil {
		return fmt.Errorf("generate NFO files: %w", err)
	}
	fmt.Println("nfo command completed successfully.")
	return nil
}
