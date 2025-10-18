package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Profile struct {
	StatusText  string `json:"status_text"`
	StatusEmoji string `json:"status_emoji"`
}

type Payload struct {
	Profile Profile `json:"profile"`
}

type Song struct {
	Title  string
	Artist string
	Album  string
}

func init() {
	log.Default().SetFlags(0)
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[slack-music-status] ")
	log.Println("starting...")

	// Check for SLACK_TOKEN env var

	if os.Getenv("SLACK_TOKEN") == "" {
		log.Fatal("please set the SLACK_TOKEN env var using a slack user token with users.profile:write scope!")
	}
}

func main() {

	lastSong := Song{}

	for {
		song := get_song()

		if song != lastSong {
			log.Println("Now playing:", song)
			update_slack_song(song)
			lastSong = song
		}

		time.Sleep(1 * time.Second)
	}
}

func get_song() Song {
	var userHomeDir, err = os.UserHomeDir()
	data, err := os.ReadFile(userHomeDir + "/music.log")
	if err != nil {
		log.Println("error reading music log:", err)
		return Song{}
	}

	lines := bytes.Split(data, []byte("\n"))
	if len(lines) < 2 {
		return Song{}
	}

	lastLine := lines[len(lines)-2] // last line is empty, so take second last
	fields := bytes.Split(lastLine, []byte("|"))

	if len(fields) < 8 {
		return Song{}
	}

	title := string(fields[5])
	artist := string(fields[6])
	album := string(fields[7])

	if title == "null" || artist == "null" || album == "null" {
		return Song{}
	}

	return Song{
		Title:  title,
		Artist: artist,
		Album:  album,
	}
}

func update_slack_song(music Song) {
	slackToken := os.Getenv("SLACK_TOKEN")

	payload := Payload{
		Profile: Profile{
			StatusText:  music.Title + " - " + music.Artist,
			StatusEmoji: ":notes:",
		},
	}

	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://slack.com/api/users.profile.set", bytes.NewBuffer(data))
	if err != nil {
		log.Println("slack request error:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+slackToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Slack request error:", err)
		return
	}
	defer resp.Body.Close()
}
