package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Nfo struct {
		RootDir       string `help:"Path to the root directory" required:""`
		FileExtension string `help:"File extension for NFO files" default:"info.json"`
		Force         bool   `help:"Force overwrite of existing NFO files" default:"false"`
	} `cmd:"" help:"Generate NFO from info.json files."`

	Download struct {
		DstDir string `help:"Path to the destination directory for downloaded files." required:""`
	} `cmd:"" help:"Downloads video files. It also downloads infojson and thumbnail files for videos and channels."`

	Rename struct {
		SrcDir string `help:"Path to the source directory for files to be renamed"`
		DstDir string `help:"Path to the destination directory for renamed files."`
		Force  bool   `help:"Force overwrite of existing files" default:"false"`
	} `cmd:"" help:"Rename video and infojson files to the channel - sxex - title.ext format."`
	// EnvFilePath string `help:"Path to the env file" default:""`
	// Verbose     bool   `help:"Enable verbose output"`
}

// <title> 	Series title
// <plot> 	Series synopsis
// <mpaa> 	Content rating
// <premiered> 	Series premiere date in YYYY-MM-DD format
// <status> 	Series status (e.g. Continuing, Ended)
// <studio> 	Network or studio
// <genre> 	Genre (multiple tags supported)
// <uniqueid> 	External ID with type attribute
// <namedseason> 	Custom season name with number attribute
// <actor> 	Cast member, containing <name>, <role>, <order>, <thumb>

type appConfig struct {
	envFilePath   string `help:"Path to the env file"`
	rootDir       string `help:"Path to the root directory"`
	fileExtension string `help:"File extension for info.json files"`
}

func NewAppConfig(env_file_path, root_dir_path, file_extension string) (appConfig, error) {
	a := appConfig{
		envFilePath:   env_file_path,
		rootDir:       root_dir_path,
		fileExtension: file_extension,
	}
	return a, nil
}

// findFiles performs resursive search from root_dir_path down for files with extension.
// Returns a string slice of complete file paths
func findFiles(root_dir_path, file_extension string) ([]string, error) {
	var matches []string

	err := filepath.WalkDir(root_dir_path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Stop walking if a directory cannot be read
		}

		// Check if it's a file and matches the target extension. Excludes apple nfo metadata files starting with "._" and other hidden files.
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), file_extension) && !strings.HasPrefix(filepath.Base(path), "._") {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
		return nil, err
	}

	return matches, nil
}

type videoInfoJson struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Channel     string   `json:"channel"`
	Channel_id  string   `json:"channel_id"`
	Channel_url string   `json:"channel_url"`
	Upload_date string   `json:"upload_date"`
	Timestamp   int64    `json:"timestamp"`
	Uploader    string   `json:"uploader"`
	Epoch       int64    `json:"epoch"`
	Tags        []string `json:"tags"`
	Categories  []string `json:"categories"`
}

// <title> 	Episode title
// <season> 	Season number
// <episode> 	Episode number
// <plot> 	Episode synopsis
// <aired> 	Air date in YYYY-MM-DD format
// <director> 	Director name
// <credits> 	Writer name
// <actor> 	Guest cast, containing <name>, <role>, <thumb>

type episodeXML struct {
	XMLName xml.Name `xml:"episodedetails"`
	Id      string   `xml:"id"`
	Title   string   `xml:"title"`
	Season  string   `xml:"season"`
	Episode string   `xml:"episode"`
	Plot    string   `xml:"plot"`
	Aired   string   `xml:"aired"`
	Genre   []string `xml:"genre"`
}

func createEpisodeNfoFile(video_infojson *videoInfoJson, nfo_file string) error {
	// timestamp := time.Unix(video_infojson.Timestamp, 0)
	// season_string := timestamp.Format("2006")      // Use the year as the season string
	// episode_string := timestamp.Format("01021504") // Use the MMDDHHMMSS as the episode string
	// aired_string := timestamp.Format("2006-01-02") // Format the aired date as YYYY-MM-DD
	// prepare to create new files with the name path and filename
	loc, _ := time.LoadLocation("America/Denver")
	uploadTime, err := time.ParseInLocation("20060102", video_infojson.Upload_date, loc)
	if err != nil {
		return fmt.Errorf("Error parsing date: %v", err)
	}
	year, month, day := uploadTime.Date()
	hour, minute, second := time.Unix(video_infojson.Timestamp, 0).In(loc).Clock()
	season_string := fmt.Sprintf("%d", int(year))                                                     // Use the year as the season string
	episode_string := fmt.Sprintf("%02d%02d%02d%02d%02d", int(month), int(day), hour, minute, second) // Use the MMDDHHMMSS as the episode string
	aired_string := uploadTime.Format("2006-01-02")                                                   // Format the aired date as YYYY-MM-DD
	safe_video_title := safeString(video_infojson.Title)

	xmlData := episodeXML{
		Id:      video_infojson.Id, // You can set this to a unique ID if needed
		Title:   safe_video_title,
		Season:  season_string,
		Episode: episode_string,
		Plot:    video_infojson.Description,
		Aired:   aired_string,
		Genre:   slices.Concat(video_infojson.Categories, video_infojson.Tags),
	}
	file, err := os.Create(nfo_file)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", " ")

	err = encoder.Encode(xmlData)
	if err != nil {
		return err
	}

	fmt.Println("XML file created successfully")
	return nil
}

type channelInfoJson struct {
	Id          string `json:"id"`          // unique channel id
	Title       string `json:"title"`       //title
	Description string `json:"description"` //plot
}

// <title> 	Series title
// <plot> 	Series synopsis
// <mpaa> 	Content rating
// <premiered> 	Series premiere date in YYYY-MM-DD format
// <status> 	Series status (e.g. Continuing, Ended)
// <studio> 	Network or studio
// <genre> 	Genre (multiple tags supported)
// <uniqueid> 	External ID with type attribute
// <namedseason> 	Custom season name with number attribute
// <actor> 	Cast member, containing <name>, <role>, <order>, <thumb>

type tvshowXML struct {
	XMLName xml.Name `xml:"tvshow"`
	Id      string   `xml:"id"`
	Title   string   `xml:"title"`
	Plot    string   `xml:"plot"`
	Studio  string   `xml:"studio"`
}

func createTVShowNfoFile(channel_infojson *channelInfoJson, nfo_file string) error {
	safe_channel_name := safeString(channel_infojson.Title)

	xmlData := tvshowXML{
		Id:     channel_infojson.Id, // You can set this to a unique ID if needed
		Title:  safe_channel_name,
		Plot:   channel_infojson.Description,
		Studio: "Youtube",
	}
	_, err := os.Stat(nfo_file)
	if err != nil {
		// nfo file does not exist, try to create the nfo file
		file, err := os.Create(nfo_file)
		if err != nil {
			return err
		}
		defer file.Close()

		encoder := xml.NewEncoder(file)
		encoder.Indent("", " ")

		err = encoder.Encode(xmlData)
		if err != nil {
			return err
		}
		fmt.Printf("XML file %s created successfully\n", nfo_file)
	}

	// Create show poster file if it does not exist.
	path := filepath.Dir(nfo_file)
	incorrect_poster_file_path := filepath.Join(path, fmt.Sprintf("%s [%s].jpg", channel_infojson.Title, channel_infojson.Id))
	correct_poster_file_path := filepath.Join(path, "show.jpg")
	_, err = os.Stat(correct_poster_file_path)
	if err != nil {
		// show.jpg file does not exist, try to create the show file
		_, err = os.Stat(incorrect_poster_file_path)
		if err == nil {
			err = copyFile(incorrect_poster_file_path, correct_poster_file_path)
			if err != nil {
				fmt.Printf("Error copying poster file %s to %s: %v\n", incorrect_poster_file_path, correct_poster_file_path, err)
				return err
			}
		}
	}
	return nil
}

func createNfoFiles(infojson_files []string, nfo_files []string, force bool) error {
	for _, infojson_file := range infojson_files {
		data, err := os.ReadFile(infojson_file)
		if err != nil {
			fmt.Printf("Error reading info.json file %s: %v\n", infojson_file, err)
			return err
		}
		if len(data) == 0 {
			fmt.Printf("Error: info.json file %s is empty\n", infojson_file)
			return fmt.Errorf("info.json file %s is empty", infojson_file)
		}
		var dataMap map[string]any
		err = json.Unmarshal(data, &dataMap)
		if err != nil {
			return err
		}
		// var id_string string
		id_string, ok := dataMap["id"].(string)
		if !ok {
			return fmt.Errorf("Failed to read infojson id field. File: %s", infojson_file)
		}
		_, hasFormats := dataMap["formats"]
		// video infojson has video formats
		if hasFormats {
			nfo_file := strings.TrimSuffix(infojson_file, ".info.json") + ".nfo"
			if slices.Contains(nfo_files, nfo_file) && force == false {
				fmt.Printf("NFO file already exists: %s. Skipping creation.\n", nfo_file)
				continue
			}
			fmt.Printf("Creating NFO file: %s\n", nfo_file)

			// Create Video file NFO file
			video_infojson := &videoInfoJson{}
			err = json.Unmarshal(data, &video_infojson)
			if err != nil {
				return err
			}
			err = createEpisodeNfoFile(video_infojson, nfo_file)
			if err != nil {
				return err
			}
			// channel infojson file have a follower count
		} else if strings.HasPrefix(id_string, "UC") {
			path, _ := filepath.Split(infojson_file)
			nfo_file := filepath.Join(path, "tvshow.nfo")
			if slices.Contains(nfo_files, nfo_file) && force == false {
				fmt.Printf("NFO file already exists: %s. Skipping creation.\n", nfo_file)
				continue
			}

			// Create tv nfo file
			channel_infojson := &channelInfoJson{}
			err = json.Unmarshal(data, &channel_infojson)
			if err != nil {
				return err
			}
			err = createTVShowNfoFile(channel_infojson, nfo_file)
			if err != nil {
				return err
			}

		} else {
			// print error if file type is not recognized. but keep going.
			// return fmt.Errorf("Unrecognized infojson file type: '%s'. id: '%s' \n", infojson_file, id_string)
			fmt.Printf("Unrecognized infojson file type: '%s'. id: '%s' \n", infojson_file, id_string)
		}
	}
	return nil
}

// Command functions
func nfo(root_dir string, file_extension string, force bool) error {
	infojson_files, err := findFiles(root_dir, file_extension)
	if err != nil {
		return fmt.Errorf("Error finding info.json files. Error: %v", err)
	}
	// Get list of existing nfo files to avoid overwriting unless necessary
	nfo_files, err := findFiles(root_dir, ".nfo")
	if err != nil {
		return fmt.Errorf("Error finding nfo files. Error: %v", err)
	}

	err = createNfoFiles(infojson_files, nfo_files, force)
	if err != nil {
		return fmt.Errorf("Error creating NFO files. Error: %v", err)
	}
	return nil
}

func findVideoFile(root_path, basename string) (string, error) {
	var match string
	video_extensions := []string{".mkv", ".mp4", ".webm"}
	err := filepath.WalkDir(root_path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Stop walking if a directory cannot be read
		}

		// Check if it's a file and matches any of the video file extensions
		if !d.IsDir() && slices.Contains(video_extensions, filepath.Ext(strings.ToLower(path))) && strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) == basename {
			match = path
			return fs.SkipDir
		} else {
			fmt.Printf("%s is not a video file or name does not container '%s'.\n", path, basename)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return match, nil

}

// copyFile copies file from the source path to the destination path and returns error.
func copyFile(src_file_path, dst_file_path string) error {
	// Open the source file
	sourceFile, err := os.Open(src_file_path)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst_file_path)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Efficiently stream the bytes from source to destination
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Flush the file system's in-memory buffers to stable storage
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}
	return nil
}

// renameFiles recursively searches src_dir and renames video files and info.json file it finds.
// Moves files to dst_dir if dst_dir is different than src_dir
func renameFiles(src_dir, dst_dir string, force bool) error {
	infojson_files, err := findFiles(src_dir, "info.json")
	if err != nil {
		return err
	}
	// Loop over infojson files and Rename video and infojson files. Optionally move the files to the dst_dir
	for _, infojson_file := range infojson_files {
		// read the infojson file
		data, err := os.ReadFile(infojson_file)
		if err != nil {
			fmt.Printf("Error reading info.json file %s: %v\n", infojson_file, err)
			return err
		}
		if len(data) == 0 {
			fmt.Printf("Error: info.json file %s is empty\n", infojson_file)
			return fmt.Errorf("info.json file %s is empty", infojson_file)
		}
		// Determine if infojson is for a video file. else skip.
		data_string := string(data)
		if strings.Contains(data_string, "https://www.youtube.com/watch?") {
			// this is a video json file. process is and its associated video file
			video_infojson := &videoInfoJson{}
			err = json.Unmarshal(data, &video_infojson)
			if err != nil {
				return err
			}
			old_path, old_infojson_name := filepath.Split(infojson_file)
			old_basename := strings.TrimSuffix(old_infojson_name, ".info.json")
			// find the video file. everything depends on it existing.
			video_file, err := findVideoFile(old_path, old_basename)
			if err != nil {
				return err
			}
			if video_file == "" {
				fmt.Printf("Failed to find video file matching %s%s. Skipping...\n", old_path, old_basename)
				continue
			}
			// Check if there is a thumbnail file. These don't exist on older youtube videos since I didn't download thumbnails in the early days.
			var new_thumbnail_file_path string
			old_thumbnail_file_path := filepath.Join(old_path, old_basename+".jpg")
			_, err = os.Stat(old_thumbnail_file_path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					old_thumbnail_file_path = ""
				} else {
					return err
				}
			}

			// check if old NFO file exists
			var new_nfo_file_path string
			old_nfo_file_path := filepath.Join(old_path, old_basename+".nfo")
			_, err = os.Stat(old_nfo_file_path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					old_nfo_file_path = ""
				} else {
					return err
				}
			}

			// prepare to create new files with the name path and filename
			loc, _ := time.LoadLocation("America/Denver")
			uploadTime, err := time.ParseInLocation("20060102", video_infojson.Upload_date, loc)
			// uploadTime, err := time.Parse("20060102", video_infojson.Upload_date)
			if err != nil {
				return fmt.Errorf("Error parsing date: %v", err)
			}
			year, month, day := uploadTime.Date()
			hour, minute, second := time.Unix(video_infojson.Timestamp, 0).In(loc).Clock()
			season_string := fmt.Sprintf("%d", int(year)) // Use the year as the season string
			season_dir_name := fmt.Sprintf("Season %d", int(year))
			episode_string := fmt.Sprintf("%02d%02d%02d%02d%02d", int(month), int(day), hour, minute, second) // Use the MMDDHHMMSS as the episode string
			video_file_ext := filepath.Ext(video_file)

			safe_channel_name := safeString(video_infojson.Channel)
			safe_video_title := safeString(video_infojson.Title)
			// Calculate new base name and path since that will be the same for all files. '/path/channel - sXeX - title [video id]'
			// Path calculation is if there is a channel directory in the path, don't add one to the path.
			// If there is a season under the channel, don't add the season.
			// Otherise , add the channel and season directories underneath the path.
			var new_dst_dir string
			if strings.Contains(dst_dir, "[UC") {
				// Patch contains channel dir. we won't add channel dir to the path.
				path_parts := strings.Split(dst_dir, "/")
				var channel_dir_index int
				// find the index of the channel dir
				for index, part := range path_parts {
					if strings.Contains(part, "[UC") {
						channel_dir_index = index
					}
				}
				re := regexp.MustCompile(`\d{4}`)
				if channel_dir_index == len(path_parts)-1 {
					// channel dir is the last item . add the season
					new_dst_dir = filepath.Join(dst_dir, season_dir_name)
				} else if (len(path_parts)-1)-channel_dir_index == 1 && re.MatchString(path_parts[len(path_parts)-1]) {
					// channel and season exist. add nothing
					new_dst_dir = dst_dir

				} else {
					return fmt.Errorf("Channel dir exists in path %s but index was not found in expected location.", dst_dir)
				}

			} else {
				// add channel and season to the path
				new_dst_dir = filepath.Join(dst_dir, fmt.Sprintf("%s [%s]", safe_channel_name, video_infojson.Channel_id), "Season "+season_string)
			}
			// create new file names
			// TODO add NFO file rename code
			new_infojson_file_path := filepath.Join(new_dst_dir, fmt.Sprintf("%s - S%sE%s - %s [%s].info.json", safe_channel_name, season_string, episode_string, safe_video_title, video_infojson.Id))
			new_video_file_path := filepath.Join(new_dst_dir, fmt.Sprintf("%s - S%sE%s - %s [%s]%s", safe_channel_name, season_string, episode_string, safe_video_title, video_infojson.Id, video_file_ext))
			if old_thumbnail_file_path != "" {
				new_thumbnail_file_path = filepath.Join(new_dst_dir, fmt.Sprintf("%s - S%sE%s - %s [%s].jpg", safe_channel_name, season_string, episode_string, safe_video_title, video_infojson.Id))
			}
			if old_nfo_file_path != "" {
				new_nfo_file_path = filepath.Join(new_dst_dir, fmt.Sprintf("%s - S%sE%s - %s [%s].nfo", safe_channel_name, season_string, episode_string, safe_video_title, video_infojson.Id))
			}

			// Check if new season dir exists before creating it
			if force == true {
				// create the channel and season directories if they don't exist
				err = os.MkdirAll(new_dst_dir, 0755)
				if err != nil {
					return fmt.Errorf("Error creating new season dir %s.: %v\n", new_dst_dir, err)
				}
			} else {
				_, err = os.Stat(new_dst_dir)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						// create the channel and season directories if they don't exist
						err = os.MkdirAll(new_dst_dir, 0755)
						if err != nil {
							return fmt.Errorf("Error creating new season dir %s.: %v\n", new_dst_dir, err)
						}
					} else {
						return err
					}
				}
			}

			if force == true {
				err = os.Rename(infojson_file, new_infojson_file_path)
				if err != nil {
					return fmt.Errorf("Error renaming infojson file '%s' to '%s': %v\n", infojson_file, new_infojson_file_path, err)
				}
				fmt.Printf("%s renamed to %s\n", infojson_file, new_infojson_file_path)
				err = os.Chtimes(new_infojson_file_path, uploadTime, uploadTime)
				if err != nil {
					return fmt.Errorf("Error changing infojson mtime file %s.: %v\n", new_infojson_file_path, err)
				}

			} else {
				// Check if new infojson file already exists
				_, err = os.Stat(new_infojson_file_path)
				// create infojson file if it does not already exist
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						err = os.Rename(infojson_file, new_infojson_file_path)
						if err != nil {
							return fmt.Errorf("Error renaming infojson file '%s' to '%s': %v\n", infojson_file, new_infojson_file_path, err)
						}
						fmt.Printf("%s renamed to %s\n", infojson_file, new_infojson_file_path)
						err = os.Chtimes(new_infojson_file_path, uploadTime, uploadTime)
						if err != nil {
							return fmt.Errorf("Error changing infojson mtime file %s.: %v\n", new_infojson_file_path, err)
						}
					} else {
						return err
					}
				}
			}
			// Video file rename section
			// Check if new video file already exists
			if force == true {
				// create video file file if it does not already exist
				err = os.Rename(video_file, new_video_file_path)
				if err != nil {
					return fmt.Errorf("Error renaming video file %s to %s: %v\n", video_file, new_video_file_path, err)
				}
				fmt.Printf("%s renamed to %s\n", video_file, new_video_file_path)
				// change the modification time of the video file to match the published timestamp
				err = os.Chtimes(new_video_file_path, uploadTime, uploadTime)
				if err != nil {
					return fmt.Errorf("Error changing video mtime file %s.: %v\n", new_video_file_path, err)
				}
			} else {
				_, err = os.Stat(new_video_file_path)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						// create video file file if it does not already exist
						err = os.Rename(video_file, new_video_file_path)
						if err != nil {
							return fmt.Errorf("Error renaming video file %s to %s: %v\n", video_file, new_video_file_path, err)
						}
						fmt.Printf("%s renamed to %s\n", video_file, new_video_file_path)
						// change the modification time of the video file to match the published timestamp
						err = os.Chtimes(new_video_file_path, uploadTime, uploadTime)
						if err != nil {
							return fmt.Errorf("Error changing video mtime file %s.: %v\n", new_video_file_path, err)
						}
					} else {
						return err
					}
				}
			}

			// thumbnail file section
			if new_thumbnail_file_path != "" {
				if force == true {
					// create thumbnail file file if it does not already exist
					err = os.Rename(old_thumbnail_file_path, new_thumbnail_file_path)
					if err != nil {
						return fmt.Errorf("Error renaming thumbnail file %s to %s: %v\n", old_thumbnail_file_path, new_thumbnail_file_path, err)
					}
					fmt.Printf("%s renamed to %s\n", old_thumbnail_file_path, new_thumbnail_file_path)
					// change the modification time of the thumbnail file to match the published timestamp
					err = os.Chtimes(new_thumbnail_file_path, uploadTime, uploadTime)
					if err != nil {
						return fmt.Errorf("Error changing the mtime on thumbnail file %s.: %v\n", new_thumbnail_file_path, err)
					}
				} else {
					_, err = os.Stat(new_thumbnail_file_path)
					if err != nil {
						if errors.Is(err, os.ErrNotExist) {
							// create thumbnail file file if it does not already exist
							err = os.Rename(old_thumbnail_file_path, new_thumbnail_file_path)
							if err != nil {
								return fmt.Errorf("Error renaming thumbnail file %s to %s: %v\n", old_thumbnail_file_path, new_thumbnail_file_path, err)
							}
							fmt.Printf("%s renamed to %s\n", old_thumbnail_file_path, new_thumbnail_file_path)
							// change the modification time of the thumbnail file to match the published timestamp
							err = os.Chtimes(new_thumbnail_file_path, uploadTime, uploadTime)
							if err != nil {
								return fmt.Errorf("Error changing the mtime on thumbnail file %s.: %v\n", new_thumbnail_file_path, err)
							}
						} else {
							return err
						}
					}
				}
			} else {
				fmt.Printf("No old thumbnail file detected at %s. Nothing to rename...\n", old_thumbnail_file_path)
			}

			// nfo file section
			if new_nfo_file_path != "" {
				if force == true {
					// create new nfo file if it does not already exist
					err = os.Rename(old_nfo_file_path, new_nfo_file_path)
					if err != nil {
						return fmt.Errorf("Error renaming nfo file %s to %s: %v\n", old_nfo_file_path, new_nfo_file_path, err)
					}
					fmt.Printf("%s renamed to %s\n", old_nfo_file_path, new_nfo_file_path)
					// change the modification time of the nfo file to match the published timestamp
					err = os.Chtimes(new_nfo_file_path, uploadTime, uploadTime)
					if err != nil {
						return fmt.Errorf("Error changing the mtime on nfo file %s.: %v\n", new_nfo_file_path, err)
					}
				} else {
					_, err = os.Stat(new_nfo_file_path)
					if err != nil {
						if errors.Is(err, os.ErrNotExist) {
							// create nfo file file if it does not already exist
							err = os.Rename(old_nfo_file_path, new_nfo_file_path)
							if err != nil {
								return fmt.Errorf("Error renaming nfo file %s to %s: %v\n", old_nfo_file_path, new_nfo_file_path, err)
							}
							fmt.Printf("%s renamed to %s\n", old_nfo_file_path, new_nfo_file_path)
							// change the modification time of the nfo file to match the published timestamp
							err = os.Chtimes(new_nfo_file_path, uploadTime, uploadTime)
							if err != nil {
								return fmt.Errorf("Error changing the mtime on nfo file %s.: %v\n", new_nfo_file_path, err)
							}
						} else {
							return err
						}
					}
				}
			} else {
				fmt.Printf("No old thumbnail file detected at %s. Nothing to rename...\n", old_thumbnail_file_path)
			}
		} else {
			// not a video infojson file. skip to next file.
			continue
		}
	}
	return nil
}

// safeString replaces characters in a string that are not allowed in file names with safe alternatives.
func safeString(input string) string {
	replacer := strings.NewReplacer(
		"/", "⧸",
		"?", "",
		":", "-",
		"*", "",
	)
	// "ä", "ae",
	// "ë", "e",
	// "ß", "ss",
	// "İ", "I",
	// "ö", "oe",
	// "ş", "s",
	// "ü", "ue",

	return replacer.Replace(input)

}

// main function orchestrates the command-line interface and dispatches commands to their respective handlers.
func main() {
	ctx := kong.Parse(&CLI)

	switch ctx.Command() {
	case "nfo":
		fmt.Println("NFO command selected.")
		err := nfo(CLI.Nfo.RootDir, CLI.Nfo.FileExtension, CLI.Nfo.Force)
		if err != nil {
			fmt.Printf("Error generating nfo files: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("nfo command completed successfully.")
	case "download":
		fmt.Println("Download command is not implemented.")
		os.Exit(1)

	case "rename":
		err := renameFiles(CLI.Rename.SrcDir, CLI.Rename.DstDir, CLI.Rename.Force)
		if err != nil {
			fmt.Printf("Error renaming files: %s", err)
			os.Exit(1)
		}
		fmt.Println("Rename command completed successfully.")
	default:
		fmt.Printf("Unknown command %s\n", ctx.Command())
		os.Exit(1)
	}
}
