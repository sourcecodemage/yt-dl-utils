package cmd

import ()

// Globals holds persistent flags available to every subcommand.
type Globals struct {
	Verbose        bool   `short:"v" help:"Enable detailed output." optional:"" default:"false"`
	ConfigFilePath string `short:"c" help:"Path to the config file" required:"" default:""`
}

type CLI struct {
	Globals
	CreateNFO               CreateNFOCmd               `cmd:"" help:"Create NFO files for videos and shows from info.json data."`
	DownloadVideos          DownloadVideosCmd          `cmd:"" help:"Download youtube videos and their metadata to the DstDir"`
	DownloadVideoMetadata   DownloadVideoMetadataCmd   `cmd:"" help:"Downloads youtube video metadata."`
	DownloadChannelMetadata DownloadChannelMetadataCmd `cmd:"" help:"Downloads youtube channel metadata."`
	RenameFiles             RenameFilesCmd             `cmd:"" help:"Rename files according to standard format and metadata in info.json file."`
}
