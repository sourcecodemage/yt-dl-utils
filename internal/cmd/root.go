package cmd

import (
	"github.com/alecthomas/kong"
)

type Globals struct {
	BatchFile              string          `help:"Path to file containing youtube video urls."`
	ChannelInventoryFile   string          `help:"Path to file containing list of youtube channel urls."`
	ChannelUnavailableFile string          `help:"Path to file containing list of unavailable youtube channel urls."`
	Config                 kong.ConfigFlag `short:"c" help:"Path to configuration file."`
	CookiesFile            string          `help:"Path to file containing netscape formatted cookies for youtube authorization."`
	CookiesFromBrowserDir  string          `help:"Path to Firefox user profile to source youtube authentication cookies from."`
	DownloadArchiveFile    string          `help:"Path to download archive file."`
	FfmpegFile             string          `help:"Path to ffmpeg binary for video and audio processing."`
	HomeDir                string          `help:"Path to the youtube library root directory."`
	VideoInventoryFile     string          `help:"Path to the youtube library inventory file."`
	ListDir                string          `help:"Path to the youtube lists folder."`
	LogDir                 string          `help:"Path to the youtube log directory."`
	TempDir                string          `help:"Path to the temporary directory for downloads."`
	VideoBlacklistFile     string          `help:"Path to file containing list of unwanted youtube video ids."`
	VideoUnavailableFile   string          `help:"Path to file containing list of unavailable youtube video ids."`
	YtdlpFile              string          `help:"Path to ffmpeg binary for video and audio processing."`
	Verbose                bool            `help:"Include debugging messages in output." default:"false"`
}

type CLI struct {
	Globals                 `embed:""`
	CreateNFO               CreateNFOCmd               `cmd:"" help:"Create NFO files for videos and shows from info.json data."`
	DownloadVideos          DownloadVideosCmd          `cmd:"" help:"Download youtube videos and their metadata to the DstDir"`
	DownloadVideoMetadata   DownloadVideoMetadataCmd   `cmd:"" help:"Downloads youtube video metadata."`
	DownloadChannelMetadata DownloadChannelMetadataCmd `cmd:"" help:"Downloads youtube channel metadata."`
	RenameFiles             RenameFilesCmd             `cmd:"" help:"Rename files according to standard format and metadata in info.json file."`
}
