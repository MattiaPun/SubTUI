package player

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"git.punjwani.pm/Mattia/DepthTUI/internal/api"
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
}

func InitPlayer() error {
	socketPath := "/tmp/depthtui_mpv_socket"

	exec.Command("pkill", "-f", socketPath).Run()
	time.Sleep(200 * time.Millisecond)

	args := []string{
		"--idle",
		"--no-video",
		"--ao=pulse",
		"--input-ipc-server=" + socketPath,
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

	return nil
}

func ShutdownPlayer() {
	if mpvCmd != nil {
		mpvCmd.Process.Kill()
	}
}

func PlaySong(songID string) error {
	if mpvClient == nil {
		return fmt.Errorf("player not initialized")
	}

	url := api.SubsonicStream(songID)
	if err := mpvClient.LoadFile(url, mpv.LoadFileModeReplace); err != nil {
		return err
	}

	api.SubsonicScrobble(songID, false)

	mpvClient.SetProperty("pause", false)

	return nil
}

func TogglePause() {
	if mpvClient == nil {
		return
	}

	status := mpvClient.IsPause()
	mpvClient.SetProperty("pause", !status)
}

func ToggleLoop() {
	if mpvClient == nil {
		return
	}

	status := mpvClient.IsPlayLoop()
	mpvClient.SetProperty("loop", !status)
}

func ToggleShuffle() {
	if mpvClient == nil {
		return
	}

	status := mpvClient.IsShuffle()
	mpvClient.SetProperty("shuffle", !status)
}

func Back10Seconds() {
	mpvClient.Seek(-10)
}

func Forward10Seconds() {
	mpvClient.Seek(+10)
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

	return PlayerStatus{
		Title:    fmt.Sprintf("%v", title),
		Artist:   fmt.Sprintf("%v", artist),
		Album:    fmt.Sprintf("%v", album),
		Current:  pos,
		Duration: dur,
		Paused:   paused,
		Volume:   vol,
	}
}
