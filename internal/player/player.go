package player

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/MattiaPun/SubTUI/internal/api"
	"github.com/gdrens/mpv"
)

var (
	mpvClient *mpv.Client
	mpvCmd    *exec.Cmd
)

type PlayerStatus struct {
	Title    string
	Artist   string
	Album    string
	Current  float64
	Duration float64
	Paused   bool
	Volume   float64
	Path     string
}

const volumeStep = 5

func InitPlayer() error {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("subtui_mpv_socket_%d", os.Getuid()))
	log.Printf("[Player] Initializing MPV IPC at %s", socketPath)

	_ = exec.Command("pkill", "-f", socketPath).Run()
	time.Sleep(200 * time.Millisecond)

	replayGain := strings.ToLower(api.AppConfig.App.ReplayGain)
	if replayGain != "track" && replayGain != "album" {
		replayGain = "no"
	}

	args := []string{
		"--idle",
		"--no-video",
		"--input-ipc-server=" + socketPath,
		"--gapless-audio=yes",
		"--prefetch-playlist=yes",
		"--replaygain=" + replayGain,
	}

	mpvCmd = exec.Command("mpv", args...)
	if err := mpvCmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %v", err)
	}

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		if _, err := os.Stat(socketPath); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	ipcc := mpv.NewIPCClient(socketPath)
	client := mpv.NewClient(ipcc)
	mpvClient = client

	log.Printf("[Player] MPV started successfully")
	return nil
}

func ShutdownPlayer() {
	if mpvCmd != nil {
		_ = mpvCmd.Process.Signal(syscall.SIGTERM)
	}
}

func PlaySong(songID string, startPaused bool) error {
	log.Printf("[Player] PlaySong called for ID: %s (Paused: %v)", songID, startPaused)

	if mpvClient == nil {
		return fmt.Errorf("player not initialized")
	}

	url := api.SubsonicStream(songID)
	if err := mpvClient.LoadFile(url, mpv.LoadFileModeReplace); err != nil {
		return err
	}

	api.SubsonicScrobble(songID, false)

	_ = mpvClient.SetProperty("pause", startPaused)

	return nil
}

func EnqueueSong(songID string) error {
	if mpvClient == nil {
		return fmt.Errorf("player not initialized")
	}

	url := api.SubsonicStream(songID)
	return mpvClient.LoadFile(url, mpv.LoadFileModeAppend)
}

func UpdateNextSong(songID string) {
	if mpvClient == nil {
		return
	}

	_ = mpvClient.PlayClear()

	if songID != "" {
		_ = EnqueueSong(songID)
	}
}

func TogglePause() {
	if mpvClient == nil {
		return
	}

	status := mpvClient.IsPause()
	_ = mpvClient.SetProperty("pause", !status)
}

func RestartSong() {
	_ = mpvClient.Seek(-int(mpvClient.Position()))

}

func Back10Seconds() {
	_ = mpvClient.Seek(-10)
}

func Forward10Seconds() {
	_ = mpvClient.Seek(+10)
}

func VolumeUp() {
	if mpvClient.CurrentVolume()+volumeStep > 100 {
		_ = mpvClient.Volume(100)
		return
	}
	_ = mpvClient.Volume(mpvClient.CurrentVolume() + volumeStep)
}

func VolumeDown() {
	if mpvClient.CurrentVolume()-volumeStep < 0 {
		_ = mpvClient.Volume(0)
		return
	}
	_ = mpvClient.Volume(mpvClient.CurrentVolume() - volumeStep)
}

func GetVolume() float64 {
	vol, _ := mpvClient.GetFloatProperty("volume")
	return vol
}

func SetVolume(volume int) {
	_ = mpvClient.Volume(volume)
}

func GetPlayerStatus() PlayerStatus {
	if mpvClient == nil {
		return PlayerStatus{}
	}

	title := mpvClient.GetProperty("media-title")
	artist := mpvClient.GetProperty("metadata/by-key/artist")
	album := mpvClient.GetProperty("metadata/by-key/album")

	pos := mpvClient.Position()
	dur := mpvClient.Duration()
	paused := mpvClient.IsPause()
	vol, _ := mpvClient.GetFloatProperty("volume")

	path := mpvClient.GetProperty("path")

	return PlayerStatus{
		Title:    fmt.Sprintf("%v", title),
		Artist:   fmt.Sprintf("%v", artist),
		Album:    fmt.Sprintf("%v", album),
		Current:  pos,
		Duration: dur,
		Paused:   paused,
		Volume:   vol,
		Path:     fmt.Sprintf("%v", path),
	}
}
