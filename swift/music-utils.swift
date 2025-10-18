import Foundation
import AppKit

class MusicObserver {
    let nc = DistributedNotificationCenter.default()
    let logURL = FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("music.log")

    var lastTrack: (name: String, artist: String, album: String)?
    var lastState: String?

    init() {
        nc.addObserver(
            forName: NSNotification.Name("com.apple.Music.playerInfo"),
            object: nil,
            queue: nil
        ) { notification in
            guard let userInfo = notification.userInfo else { return }

            let name = userInfo["Name"] as? String ?? "null"
            let artist = userInfo["Artist"] as? String ?? "null"
            let album = userInfo["Album"] as? String ?? "null"
            let state = userInfo["Player State"] as? String ?? "null"
            let currTrack = (name, artist, album)

            var action = "UNKNOWN"

            if let prev = self.lastTrack {
                if currTrack != prev {
                    if self.lastState == "Playing" && state == "Playing" {
                        action = "NEXT"
                    } else {
                        action = "TRACK_CHANGE"
                    }
                } else if self.lastState != state {
                    switch state {
                    case "Playing": action = "PLAY"
                    case "Paused": action = "PAUSE"
                    case "Stopped": action = "STOP"
                    default: action = state.uppercased()
                    }
                }
            } else {
                if state == "Playing" {
                    action = "PLAY"
                }
            }

            if action != "UNKNOWN" {
                let prev = self.lastTrack ?? ("null", "null", "null")
                let entry = "\(Date().timeIntervalSince1970)|\(action)|\(prev.0)|\(prev.1)|\(prev.2)|\(currTrack.0)|\(currTrack.1)|\(currTrack.2)\n"
                if let data = entry.data(using: .utf8) {
                    if FileManager.default.fileExists(atPath: self.logURL.path) {
                        if let handle = try? FileHandle(forWritingTo: self.logURL) {
                            handle.seekToEndOfFile()
                            handle.write(data)
                            handle.closeFile()
                        }
                    } else {
                        try? data.write(to: self.logURL)
                    }
                }
            }

            self.lastTrack = currTrack
            self.lastState = state
        }
    }
}

let observer = MusicObserver()
RunLoop.current.run()
