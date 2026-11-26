package usecase

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

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

// createEpisodeId create hash of video id , truncates it to 5 digits, and returns it and an error.
func createEpisodeId(id string) (int, error) {
	h := fnv.New32a()
	h.Write([]byte(id))
	// Map the 32-bit hash into the range [10000, 99999]
	return int(h.Sum32()%90000) + 10000, nil

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
	fmt.Printf("Creating Episode NFO file: %s\n", nfo_file)

	// prepare to create new files with the name path and filename
	loc, _ := time.LoadLocation("America/Denver")
	uploadTime, err := time.ParseInLocation("20060102", video_infojson.Upload_date, loc)
	if err != nil {
		return fmt.Errorf("Error parsing date: %v", err)
	}
	year, month, day := uploadTime.Date()
	// hour, minute, second := time.Unix(video_infojson.Timestamp, 0).In(loc).Clock()
	season_string := fmt.Sprintf("%d", int(year)) // Use the year as the season string
	// generate a unique episode id from the video id. This id is for ordering videos that were uploaded on the same date. Since upload has a date but not a time, we have to make up an determinstic values.
	// in Plex, the episode field is limited to 9 chars. So episode is MMDD + 5 digit episode id.
	episode_id, err := createEpisodeId(video_infojson.Id)
	if err != nil {
		return fmt.Errorf("Error generating episode id: %v", err)
	}
	episode_string := fmt.Sprintf("%02d%02d%05d", int(month), int(day), episode_id) // Use the MMDDHHMMSS as the episode string
	aired_string := uploadTime.Format("2006-01-02")                                 // Format the aired date as YYYY-MM-DD
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

	fmt.Printf("Created Episode NFO file: %s succcessfully.\n", nfo_file)
	return nil
}

// createEpisodeNfoFiles uses info.json file to generate episode information and create an episode nfo file. Returns error if there is an issue.
func createEpisodeNfoFiles(root_dir string, force bool) error {
	// find video info json files
	infojson_files, err := findFiles(root_dir, `^.+\s\[\w{11}\]\.info\.json$`)
	if err != nil {
		return fmt.Errorf("Error finding info.json files. Error: %v", err)
	}
	err = nil

	// Get list of existing video nfo files to avoid overwriting unless necessary
	nfo_files, err := findFiles(root_dir, `^.+\s\[\w{11}\]\.nfo$`)
	if err != nil {
		return fmt.Errorf("Error finding nfo files. Error: %v", err)
	}

	// Find all video files
	video_files, err := findFiles(root_dir, ".mkv")
	if err != nil {
		return fmt.Errorf("Error finding video files. Error: %v", err)
	}

	for _, infojson_file := range infojson_files {
		// make sure there is a matching video file.
		infojson_basename := filepath.Base(infojson_file)
		infojson_dir := filepath.Dir(infojson_file)
		idx := strings.IndexByte(infojson_basename, '.')
		infojson_rootname := infojson_basename[:idx]
		infojson_videofile_path := filepath.Join(infojson_dir, infojson_rootname+".mkv")
		if !slices.Contains(video_files, infojson_videofile_path) {
			continue
		}
		// read the video infojson file
		data, err := os.ReadFile(infojson_file)
		if err != nil {
			fmt.Printf("Error reading info.json file %s: %v\n", infojson_file, err)
			return err
		}
		if len(data) == 0 {
			return fmt.Errorf("info.json file %s is empty", infojson_file)
		}
		// create struct containing video infojson data
		video_infojson := videoInfoJson{}
		err = json.Unmarshal(data, &video_infojson)
		if err != nil {
			return err
		}
		nfo_file := strings.TrimSuffix(infojson_file, ".info.json") + ".nfo"
		if slices.Contains(nfo_files, nfo_file) && force == false {
			fmt.Printf("NFO file already exists: %s. Skipping creation.\n", nfo_file)
			continue
		}

		// Create episode NFO file from video_infojson struct
		err = createEpisodeNfoFile(&video_infojson, nfo_file)
		if err != nil {
			return err
		}
	}
	return nil
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
type channelInfoJson struct {
	Id          string `json:"id"`          // unique channel id
	Title       string `json:"title"`       //title
	Description string `json:"description"` //plot
}

type showXML struct {
	XMLName xml.Name `xml:"tvshow"`
	Id      string   `xml:"id"`
	Title   string   `xml:"title"`
	Plot    string   `xml:"plot"`
	Studio  string   `xml:"studio"`
}

func createShowNfoFile(channel_infojson *channelInfoJson, nfo_file string) error {
	fmt.Printf("Creating Show NFO file '%s'.\n", nfo_file)
	xmlData := showXML{
		Id:     channel_infojson.Id, // You can set this to a unique ID if needed
		Title:  channel_infojson.Title,
		Plot:   channel_infojson.Description,
		Studio: "Youtube",
	}
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
	fmt.Printf("Created Show NFO file '%s' successfully.\n", nfo_file)

	return nil
}

func createShowNfoFiles(root_dir string, force bool) error {
	// find show info json files
	infojson_files, err := findFiles(root_dir, `^.+\s\[UC\w{22}\]\.info\.json$`)
	if err != nil {
		return fmt.Errorf("Error finding show info.json files. Error: %v", err)
	}
	err = nil

	// Get list of existing tvshow nfo files to avoid overwriting unless necessary
	nfo_files, err := findFiles(root_dir, `^tvshow\.nfo$`)
	if err != nil {
		return fmt.Errorf("Error finding show nfo files. Error: %v", err)
	}

	if force == false {
		infojson_files = slices.DeleteFunc(infojson_files, func(s string) bool {
			dir := filepath.Dir(s)
			newpath := filepath.Join(dir, "tvshow.nfo")
			return slices.Contains(nfo_files, newpath)
		})
	}

	for _, infojson_file := range infojson_files {
		// read the show infojson file
		data, err := os.ReadFile(infojson_file)
		if err != nil {
			return fmt.Errorf("Error reading show info.json file %s: %v\n", infojson_file, err)
		}
		if len(data) == 0 {
			return fmt.Errorf("show info.json file %s is empty", infojson_file)
		}
		// create path for show nfo file
		path := filepath.Dir(infojson_file)
		nfo_file := filepath.Join(path, "tvshow.nfo")

		channel_infojson := &channelInfoJson{}
		err = json.Unmarshal(data, &channel_infojson)
		if err != nil {
			return err
		}
		// Create tv show nfo file
		err = createShowNfoFile(channel_infojson, nfo_file)
		if err != nil {
			return err
		}
	}
	return nil
}

func createShowThumbnailFile(file_path string) error {
	// get the parent dir of the channel thumbnail files
	dir := filepath.Dir(file_path)
	// determine the show thumbnail file path
	show_thumbnail_file := filepath.Join(dir, "show.jpg")
	err := copyFile(file_path, show_thumbnail_file)
	if err != nil {
		return fmt.Errorf("Error copying channel thumbnail file from %s to %s: %v\n", file_path, show_thumbnail_file, err)
	}
	return nil
}

func createShowThumbnailFiles(root_dir string, force bool) error {
	// Find channel thumbnail files which we'll copy to show.jpg
	channel_thumbnail_files, err := findFiles(root_dir, `^.+\s\[UC\w{22}\]\.jpg$`)
	if err != nil {
		return fmt.Errorf("Error finding channel thumbnail files. Error: %v", err)
	}
	err = nil

	// Get list of existing show thumbnail files to avoid overwriting unless necessary
	show_thumbnail_files, err := findFiles(root_dir, `^show\.jpg$`)
	if err != nil {
		return fmt.Errorf("Error finding show thumbnail files files. Error: %v", err)
	}
	// remove channel thumbnail files from list if the show.jpg already exists
	if force == false {
		channel_thumbnail_files = slices.DeleteFunc(channel_thumbnail_files, func(s string) bool {
			dir := filepath.Dir(s)
			newpath := filepath.Join(dir, "show.jpg")
			return slices.Contains(show_thumbnail_files, newpath)
		})
	}
	for _, channel_thumbnail_file := range channel_thumbnail_files {
		err = createShowThumbnailFile(channel_thumbnail_file)
		if err != nil {
			return fmt.Errorf("Error creating thumbnail file. Error: %v", err)
		}
	}
	return nil
}

type NFOCreator struct {
	// Inject interfaces here (e.g., database or API client)
}

func (u *NFOCreator) Execute(root_dir string, force bool) error {

	// Create episode NFO files
	err := createEpisodeNfoFiles(root_dir, force)
	if err != nil {
		return fmt.Errorf("Error creating Episode NFO files. Error: %v", err)
	}
	// Create episode NFO files
	err = createShowNfoFiles(root_dir, force)
	if err != nil {
		return fmt.Errorf("Error creating Show NFO files. Error: %v", err)
	}

	// Create show thumbnail filesfiles
	err = createShowThumbnailFiles(root_dir, force)
	if err != nil {
		return fmt.Errorf("Error creating Show thumbnail files. Error: %v", err)
	}
	return nil
}
