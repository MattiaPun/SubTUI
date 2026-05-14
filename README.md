# SubTUI

SubTUI is your next favorite lightweight, terminal-based music player for Subsonic-compatible servers like Navidrome, Gonic, and Airsonic. Built with Go and the Bubble Tea framework, it provides a clean terminal interface to listen to your favorite high-quality audio.

## Key Features

* **Subsonic-compatible**: Connect and stream from any Subsonic-compatible server
* **Format Compatibility**: Uses [mpv](https://mpv.io/) to support various audio codecs and reliable playback
* **Fully Customizable**: Configure keybinds, color themes, and settings via a simple TOML file
* **ReplayGain Support**: Built-in support for Track and Album volume normalization
* **Scrobbling**: Automatically updates your play counts on your server and external services like Last.FM or ListenBrainz
* **Gapless Playback**: Enjoy your favorite albums exactly as intended with smooth, uninterrupted transitions
* **MPRIS Support**: Control SubTUI from any media widget on Linux/FreeBSD
* **Discord Integration**: Show off what you're listening to with built-in Discord Rich Presence

![Main View](./screenshots/main_view.png)

## Installation

You must have [mpv](https://mpv.io/) installed and available in your system `PATH`. You can verify with `mpv --version`.

Pre-compiled binaries for Linux and macOS are available on the [Releases](https://github.com/MattiaPun/SubTUI/releases) page.

| Method               	   | Command / Instructions                                                         	|
|----------------------	   |--------------------------------------------------------------------------------	|
| **Debian / Ubuntu**  	   | Download the `.deb` and run `sudo dpkg -i subtui_*.deb`                        	|
| **Fedora / RHEL**    	   | Download the `.rpm` and run `sudo rpm -i subtui_*.rpm`                         	|
| **Alpine**           	   | Download the `.apk` and run `sudo apk add --allow-untrusted ./subtui_*.apk`    	|
| **Arch Linux (AUR)** 	   | `yay -S subtui-git`                                                            	|
| **macOS (Homebrew)** 	   | `brew install MattiaPun/subtui/subtui`                                         	|
| **FreeBSD**          	   | `pkg install subtui`                                                           	|
| **Nix**              	   | `nix profile install github:MattiaPun/SubTUI`                                  	|
| **Void Linux**           | `xbps-install SubTUI`                                                            |
| **Gentoo Linux (GURU)**  | `eselect repository enable guru && emerge -av media-sound/subtui`                |
| **Go Toolchain**     	   | `go install github.com/MattiaPun/SubTUI@latest`                                	|
| **From Source**      	   | `git clone https://github.com/MattiaPun/SubTUI.git && cd SubTUI && go build .` 	|

## Documentation

For setup, configuration, keybinds, and more, check out the **[Wiki](https://github.com/MattiaPun/SubTUI/wiki)**.

## Screenshots

![Login](./screenshots/login.png)
![Queue](./screenshots/queue_view.png)
![Media Player](./screenshots/media_player.png)

## Contributing

Contributions are welcome! There are several ways to help:

- **Feature Requests** — Open an [issue](https://github.com/MattiaPun/SubTUI/issues) to suggest new features or improvements
- **Code Contributions** — Fork the repo, make your changes, and submit a pull request. Please use [Conventional Commit Messages](https://www.conventionalcommits.org/en/v1.0.0/)

## Sponsor

If you enjoy using SubTUI, please consider [sponsoring](https://github.com/sponsors/MattiaPun) the project to support its development.

## License

Distributed under the MIT License. See `LICENSE` for more information.
