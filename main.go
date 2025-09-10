package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Profile struct {
	StatusText  string `json:"status_text"`
	StatusEmoji string `json:"status_emoji"`
}

type Payload struct {
	Profile Profile `json:"profile"`
}

func init() {
	log.Default().SetFlags(0)
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[slack-music-status] ")
	log.Println("starting...")

	// Detect os

	switch runtime.GOOS {
	case "darwin":
		log.Println("macos user, only apple music is supported for now :)")
	case "linux":
		log.Println("woa linux :D, install playerctl")
	default:
		log.Fatal("i think we don't support your os yet : " + runtime.GOOS)
	}

	// Check for SLACK_TOKEN env var

	if os.Getenv("SLACK_TOKEN") == "" {
		log.Fatal("please set the SLACK_TOKEN env var using a slack user token with users.profile:write scope!")
	}
}

func main() {

	lastSong := ""

	for {
		song := get_song()

		if song != lastSong {
			if song == "" {
				log.Println("Nothing playing, clearing status…")
				update_slack_song("")
			} else {
				log.Println("Now playing:", song)
				update_slack_song(song)
			}
			lastSong = song
		}

		time.Sleep(5 * time.Second)
	}
}

func get_song() string {
	switch runtime.GOOS {
	case "darwin":
		script := `tell application "Music" to if player state is playing then artist of current track & " - " & name of current track`
		cmd := exec.Command("osascript", "-e", script)
		output, err := cmd.Output()
		if err != nil {
			log.Println("error:", err)
			return ""
		}

		song := strings.TrimSpace(string(output))

		return song
	case "linux":
		cmd := exec.Command("playerctl", "metadata", "--format", "{{artist}} - {{title}}")
		output, err := cmd.Output()
		if err != nil {
			return "something went wrong, maybe you need to install playerctl?"
		}
		return strings.TrimSpace(string(output))

	default:
		// another os?
		return "something went wrong..."
	}
}

func update_slack_song(music string) {
	slackToken := os.Getenv("SLACK_TOKEN")

	payload := Payload{
		Profile: Profile{
			StatusText:  music,
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
