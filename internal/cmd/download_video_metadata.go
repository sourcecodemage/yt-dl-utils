package cmd

// DownloadVideoMetadataCmd defines the flags and positional arguments for `my-cli download-video-metadata`.
type DownloadVideoMetadataCmd struct {
	RootDir string `arg:"" help:"Path to the root directory" required:""`
	Force   bool   `help:"Force overwrite of existing NFO files" default:"false"`
}

// Run is automatically invoked by Kong when the create-nfo command is selected.
func (d *DownloadVideoMetadataCmd) Run(globals *Globals) error {
	return nil
}
