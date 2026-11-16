package config

import (
	"encoding/json"
	"os"
)

type AppConfig struct {
	BatchFilePath              string `help:"Path to file containing youtube video urls."`
	ChannelListPath            string `help:"Path to file containing list of youtube channel urls."`
	ChannelUnavailableFilePath string `help:"Path to file containing list of unavailable youtube channel urls."`
	CookiesFilePath            string `help:"Path to file containing netscape formatted cookies for youtube authorization."`
	CookiesFromBrowser         string `help:"Path to Firefox user profile to source youtube authentication cookies from."`
	DownloadArchiveFilePath    string `help:"Path to download archive file."`
	FfmpegLocation             string `help:"Path to ffmpeg binary for video and audio processing."`
	HomeDirPath                string `help:"Path to the youtube library root directory."`
	VideoInventoryFilePath     string `help:"Path to the youtube library inventory file."`
	ListDirPath                string `help:"Path to the youtube lists folder."`
	LogDirPath                 string `help:"Path to the youtube log directory."`
	TempDirPath                string `help:"Path to the temporary directory for downloads."`
	VideoBlacklistFilePath     string `help:"Path to file containing list of unwanted youtube video ids."`
	VideoUnavailableFilePath   string `help:"Path to file containing list of unavailable youtube video ids."`
	YtdlpFilePath              string `help:"Path to ffmpeg binary for video and audio processing."`
}

func (a *AppConfig) New(config_file_path string) error {
	configFile, err := os.Open(config_file_path)
	if err != nil {
		return err
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(a)
	return nil
}
