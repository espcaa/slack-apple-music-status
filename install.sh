#!/bin/bash

data_dir="$HOME/.apple-music-thingies"
launch_agents_dir="$HOME/Library/LaunchAgents"

# remove
if [ "$1" = "remove" ]; then
    echo "removing $data_dir..."
    rm -rf "$data_dir"
    echo "removing launch agents..."
    rm -f "$launch_agents_dir/com.user.apple-music-thingies.plist"
    rm -f "$launch_agents_dir/com.user.slack-apple-music-status.plist"
    echo "my code is removed forever from your mac :D"
    exit 0
fi

if [ "$1" = "update" ]; then
    echo "updating the music observer utility..."
    swiftc -o "$data_dir/music-logger" ./swift/music-utils.swift
    chmod +x "$data_dir/music-logger"
    # reload the launch agent
    launchctl bootout gui/$(id -u) "$launch_agents_dir/com.user.apple-music-thingies.plist"
    launchctl bootstrap gui/$(id -u) "$launch_agents_dir/com.user.apple-music-thingies.plist"
    echo "music observer utility updated!"

    # now the slack utility
    if [ -f "$data_dir/slack" ]; then
        echo "updating the slack status utility..."
        go build -o "$data_dir/slack" .
        chmod +x "$data_dir/slack"
        # reload the launch agent
        launchctl bootout gui/$(id -u) "$launch_agents_dir/com.user.slack-apple-music-status.plist"
        launchctl bootstrap gui/$(id -u) "$launch_agents_dir/com.user.slack-apple-music-status.plist"
        echo "slack status utility updated!"
    fi
    exit 0
fi

# create data directory
echo "creating data directory at $data_dir..."
mkdir -p "$data_dir"

# build the swift utility
echo "building the swift utility..."
swiftc -o "$data_dir/music-logger" ./swift/music-utils.swift
chmod +x "$data_dir/music-logger"

# ask to run at startup
read -rp "do you want the music observer to run at startup? (y/n) " startup_choice
if [[ "$startup_choice" =~ ^[Yy]$ ]]; then
    echo "creating the music observer plist..."
    mkdir -p "$launch_agents_dir"
    cat > "$launch_agents_dir/com.user.apple-music-thingies.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.apple-music-thingies</string>
    <key>ProgramArguments</key>
    <array>
        <string>$data_dir/music-logger</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$data_dir/musicobserver.log</string>
    <key>StandardErrorPath</key>
    <string>$data_dir/musicobserver.err</string>
</dict>
</plist>
EOF
    echo "enabling the music observer service..."
    launchctl bootstrap gui/$(id -u) "$launch_agents_dir/com.user.apple-music-thingies.plist"
fi

# slack integration
read -rp "do you want to install the Slack status integration now? (y/n) " slack_choice
if [[ "$slack_choice" =~ ^[Yy]$ ]]; then
    echo "building the slack go utility..."
    if ! command -v go &> /dev/null; then
        echo "go is not installed. please install go and re-run the installer."
        exit 1
    fi
    go build -o "$data_dir/slack" .
    chmod +x "$data_dir/slack"

    # ask to run slack integration at startup
    read -rp "do you want the slack status integration to run at startup? (y/n) " slack_startup_choice
    if [[ "$slack_startup_choice" =~ ^[Yy]$ ]]; then
        echo "creating the slack plist..."
        cat > "$launch_agents_dir/com.user.slack-apple-music-status.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.slack-apple-music-status</string>
    <key>ProgramArguments</key>
    <array>
        <string>$data_dir/slack</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$data_dir/slack.log</string>
    <key>StandardErrorPath</key>
    <string>$data_dir/slack.err</string>
</dict>
</plist>
EOF
        echo "enabling the slack service..."
        launchctl bootstrap gui/$(id -u) "$launch_agents_dir/com.user.slack-apple-music-status.plist"
    fi

    echo "note: you need to export a Slack user token as SLACK_TOKEN in order for the slack status integration to work."
fi

echo "installation complete!"
