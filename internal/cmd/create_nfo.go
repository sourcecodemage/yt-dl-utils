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
func (cmd *CreateNFOCmd) Run(cli *CLI) error {

	fmt.Println("Create NFO command selected.")
	creator := &usecase.NFOCreator{}
	if err := creator.Execute(cli.HomeDir, cmd.Force); err != nil {
		return fmt.Errorf("generate NFO files: %w", err)
	}
	fmt.Println("nfo command completed successfully.")
	return nil
}
