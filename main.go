package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func main() {
	lastSong := ""

	for {
		song := get_song()

		if song != lastSong {
			if song == "" {
				fmt.Println("Nothing playing, clearing status…")
				update_slack_song("")
			} else {
				fmt.Println("Now playing:", song)
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
			fmt.Println("Error:", err)
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
	if slackToken == "" {
		fmt.Println("Missing SLACK_TOKEN env var")
		return
	}

	payload := Payload{
		Profile: Profile{
			StatusText:  music,
			StatusEmoji: ":notes:",
		},
	}

	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://slack.com/api/users.profile.set", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Slack request error:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+slackToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Slack request error:", err)
		return
	}
	defer resp.Body.Close()
}
