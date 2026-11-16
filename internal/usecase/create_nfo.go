package usecase

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
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

type NFOCreator struct {
	// Inject interfaces here (e.g., database or API client)
}

func (u *NFOCreator) Execute(root_dir string, force bool) error {
	infojson_files, err := findFiles(root_dir, "info.json")
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
