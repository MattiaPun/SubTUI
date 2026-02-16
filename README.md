# SubTUI

SubTUI is your next favorite lightweight music player for Subsonic-compatible servers like Navidrome, Gonic, and Airsonic. Built with Go and the Bubble Tea framework, it provides a clean terminal interface to listen to your favorite high-quality audio.

## Key Features
* **Subsonic-compatible**: Connect and stream from any Subsonic-compatible server
* **Format comaptiblity**: Uses `mpv` to support various audio codecs and reliable playback
* **Fully Customizable**: Configure keybinds, color themes, and settings via a simple TOML file
* **ReplayGain Support**: Built-in support for Track and Album volume normalization
* **Scrobbling**: Automatically updates your play counts on your server and external services like Last.FM or ListenBrainz
* **Gapless Playback**: Enjoy your favorite albums exactly as intented with smooth, uninterrupted transitions
* **Discord Integrations**: Show of your listing to with built-in Discord Rich Presence

![Main View](./screenshots/main_view.png)

## Installation

### Prerequisites

You must have **mpv** installed and available in your system path.

* **Ubuntu/Debian:** `sudo apt install mpv`
* **Arch:** `sudo pacman -S mpv`
* **macOS:** `brew install mpv`

### Linux & macOS (Releases)

You can download pre-compiled binaries for Linux and macOS directly from the [Releases](https://github.com/MattiaPun/SubTUI/releases) page.

* **Debian/Ubuntu**: Download the .deb file and run `sudo dpkg -i subtui_*.deb`
* **Fedora/RHEL**: Download the .rpm file and run `sudo rpm -i subtui_*.rpm`
* **Alpine**: Download the .apk file and run `sudo apk add --allow-untrusted ./subtui_*.apk`
* **Generic**: Download the release for your architecture, extract it, and run the binary.

### macOS (Homebrew)

You can install SubTUI via the official Homebrew tap `brew install MattiaPun/subtui/subtui`

### Arch Linux (AUR)

You can install SubTUI directly from the AUR: `yay -S subtui-git`

### FreeBSD

You can install SubTUI directly via `pkg`: `pkg install subtui`
Note that this will automatically install the `mpv` dependency

### Nix

You can install SubTUI using the flake: `nix profile install github:MattiaPun/SubTUI`

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

## Configuration
On the first launch, SubTUI will generate default configuration files at: `~/.config/subtui/config.toml` and `~/.config/subtui/credentials.toml`.
 **Security Note**: Your credentials are stored in plaintext in the `credentials.toml` file

You can edit these files to save your credentials, change the color theme, or remap any keybind. You can find the default configuration templates in the repository at [internal/api/config.toml](internal/api/config.toml) and [internal/api/credentials.toml](interlan/api/credentials.toml)

## Default keybinds
**Note**: All keybinds below are the defaults. You can customize them in your config.toml.

### Global Navigation

| Key             | Action                                                 |
| --------------- | ------------------------------------------------------ |
| `Tab`           | Cycle focus forward (Search → Sidebar → Main → Footer) |
| `Shift` + `Tab` | Cycle focus backward                                   |
| `Enter`         | Play selection / Open Album                            |
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
| `A`     | Added selection to playlist |
| `R`     | Added rating to selection   |
| `G`     | Move selection to bottom    |
| `gg`    | Move selection to top       |
| `ga`    | Go to album of selection    |
| `gr`    | Go to artist of selection   |

### Media Controls

| Key       | Action                                   |
| --------- | ---------------------------------------- |
| `p` / `P` | Toggle play/pause                        |
| `n`       | Play next song                           |
| `b`       | Play previous song                       |
| `S`       | Shuffle Queue (Keeps current song first) |
| `L`       | Toggle Loop (None → All → One)           |
| `w`       | Restart song                             |
| `,`       | Rewind 10 seconds                        |
| `;`       | Forward 10 seconds                       |
| `v`       | Volume Up (+5%)                          |
| `V`       | Volume down (-5%)                        |

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

| Key        | Action                |
|------------|-----------------------|
| `s`        | Toggle notifications  |
| `Ctrl + s` | Create shareable link |


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

## Sponsor

If you enjoy using SubTUI, please consider [sponsoring](https://github.com/sponsors/MattiaPun) the project to support its development.

## License

Distributed under the MIT License. See `LICENSE` for more information.
