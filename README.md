# SubTUI

**SubTUI** is a lightweight TUI music player for Subsonic-compatible servers (Navidrome, Gonic, Airsonic, etc.) built with Go and the Bubble Tea framework. It uses `mpv` as the underlying audio engine supporting multiple audio formats. It supports scrobbling ensuring your play counts are updated on your server and on any external services configured like Last.FM or ListenBrainz. It also features Discord Rich Presence integration, displaying your current track and status on your profile.

![Main View](./screenshots/main_view.png)

## Installation

### Prerequisites

You must have **mpv** installed and available in your system path.

- **Ubuntu/Debian:** `sudo apt install mpv`
- **Arch:** `sudo pacman -S mpv`
- **macOS:** `brew install mpv`

### From Releases

You can download pre-compiled binaries for Linux and macOS directly from the [Releases](https://github.com/MattiaPun/SubTUI/releases) page. Simply download the archive for your architecture, extract it, and run the binary.

### Arch Linux (AUR)

You can install SubTUI directly from the AUR: `yay -S subtui-git`

### GoLang Toolchain

You can install SubTUI directly using GoLang: `go install github.com/MattiaPun/SubTUI@latest`

### From Source

```bash
# Clone the repo
git clone https://github.com/MattiaPun/SubTUI.git
cd SubTUI

# Build
go build .

# Run
./subtui
```

## Keybinds

### Global Navigation

| Key             | Action                                                 |
| --------------- | ------------------------------------------------------ |
| `Tab`           | Cycle focus forward (Search → Sidebar → Main → Footer) |
| `Shift` + `Tab` | Cycle focus backward                                   |
| `Backspace`     | Back                                                   |
| `?`             | Toggle help menu                                       |
| `j` / `Down`    | Move selection down                                    |
| `k` / `Up`      | Move selection up                                      |
| `q`             | Quit application (except during Login)                 |
| `Ctrl` + `c`    | Quit application                                       |

### Search

| Key          | Action                                         |
| ------------ | ---------------------------------------------- |
| `/`          | Focus the Search bar                           |
| `Ctrl` + `n` | Cycle filter forward (Songs → Albums → Artist) |
| `Ctrl` + `b` | Cycle filter backward                          |

### Library & Playlists

| Key     | Action                      |
| ------- | --------------------------- |
| `G`     | Move selection to bottom    |
| `gg`    | Move selection to top       |
| `ga`    | Go to album of selection    |
| `gr`    | Go to artist of selection   |
| `Enter` | Play selection / Open Album |

### Media Controls

| Key       | Action                                   |
| --------- | ---------------------------------------- |
| `p` / `P` | Toggle play/pause                        |
| `n`       | Play next song                           |
| `b`       | Play previous song                       |
| `Enter`   | Play selection / Open Album              |
| `S`       | Shuffle Queue (Keeps current song first) |
| `L`       | Toggle Loop (None → All → One)           |
| `w`       | Restart song                             |
| `,`       | Rewind 10 seconds                        |
| `;`       | Forward 10 seconds                       |

### Starred (liked) songs

| Key | Action             |
| --- | ------------------ |
| `f` | Toggle star        |
| `F` | Open starred Songs |

### Queue Management

| Key | Action                   |
| --- | ------------------------ |
| `Q` | Toggle queue             |
| `N` | Queue next               |
| `a` | Queue last               |
| `d` | Remove song from queue   |
| `D` | Clear queue              |
| `K` | Move song up (Reorder)   |
| `J` | Move song down (Reorder) |

### Other

| Key | Action               |
| --- | -------------------- |
| `s` | Toggle notifications |

## Configuration

On the first launch, SubTUI will ask for your server credentials:

1. **Server URL:** (e.g., `http(s)://music.example.com`)
2. **Username**
3. **Password**

**Security Note**: Your credentials are stored in plaintext in `~/.config/subtui/config.yaml`.

## Screenshots

![Login](./screenshots/login.png)
![Queue](./screenshots/queue_view.png)

## Contributing

Contributions are welcome!
Please make use of [Convention Commit Messages](https://www.conventionalcommits.org/en/v1.0.0/)

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.
