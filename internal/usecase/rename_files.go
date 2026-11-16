package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

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

type RenameFilesCreator struct {
	// Inject interfaces here (e.g., database or API client)
}

func (r *RenameFilesCreator) Execute(src_dir string, dst_dir string, force bool) error {
	err := renameFiles(src_dir, dst_dir, force)
	if err != nil {
		fmt.Printf("Error renaming files: %s", err)
		os.Exit(1)
	}
	fmt.Println("RenameFiles command completed successfully.")
	return nil
}
