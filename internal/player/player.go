package player

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
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

const (
	volumeStep          = 5
	mpvSocketTimeout    = 5 * time.Second
	mpvSocketPollDelay  = 100 * time.Millisecond
	mpvDialProbeTimeout = 200 * time.Millisecond
)

func waitForMpvSocket(socketPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for {
		conn, err := net.DialTimeout("unix", socketPath, mpvDialProbeTimeout)
		if err == nil {
			_ = conn.Close()
			return nil
		}

		lastErr = err
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for mpv IPC socket %s: %w", socketPath, lastErr)
		}

		time.Sleep(mpvSocketPollDelay)
	}
}

func stopMpvProcess() {
	if mpvCmd == nil || mpvCmd.Process == nil {
		return
	}

	_ = mpvCmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		done <- mpvCmd.Wait()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = mpvCmd.Process.Kill()
		<-done
	}
}

func InitPlayer() error {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("subtui_mpv_socket_%d", os.Getuid()))
	log.Printf("[Player] Initializing MPV IPC at %s", socketPath)

	// Clean up stale socket if one exists from a previous crash.
	_ = os.Remove(socketPath)

	killArg := fmt.Sprintf("--input-ipc-server=%s", socketPath)
	_ = exec.Command("pkill", "-f", "--", killArg).Run()

	replayGain := strings.ToLower(api.AppConfig.App.ReplayGain)
	if replayGain != "track" && replayGain != "album" {
		replayGain = "no"
	}

	gaplessPlayback := strings.ToLower(api.AppConfig.App.GaplessPlayBack)
	if gaplessPlayback != "no" && gaplessPlayback != "weak" {
		gaplessPlayback = "yes"
	}

	args := []string{
		"--idle",
		"--no-video",
		"--input-ipc-server=" + socketPath,
		"--gapless-audio=" + gaplessPlayback,
		"--prefetch-playlist=yes",
		"--replaygain=" + replayGain,
	}

	mpvCmd = exec.Command("mpv", args...)
	if err := mpvCmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %v", err)
	}

	if err := waitForMpvSocket(socketPath, mpvSocketTimeout); err != nil {
		stopMpvProcess()
		return err
	}

	ipcc := mpv.NewIPCClient(socketPath)
	client := mpv.NewClient(ipcc)
	mpvClient = client

	log.Printf("[Player] MPV started successfully")
	return nil
}

func ShutdownPlayer() {
	stopMpvProcess()
	mpvClient = nil
	mpvCmd = nil
}

func PlaySong(songID string, startPaused bool) error {
	log.Printf("[Player] PlaySong called for ID: %s (Paused: %v)", songID, startPaused)

	if mpvClient == nil {
		return fmt.Errorf("player not initialized")
	}

	url := api.SubsonicStream(songID) + fmt.Sprintf("&_nonce=%d", time.Now().UnixNano())
	if err := mpvClient.LoadFile(url, mpv.LoadFileModeReplace); err != nil {
		return err
	}

	_ = mpvClient.SetProperty("pause", startPaused)

	return nil
}

func EnqueueSong(songID string) error {
	if mpvClient == nil {
		return fmt.Errorf("player not initialized")
	}

	url := api.SubsonicStream(songID) + fmt.Sprintf("&_nonce=%d", time.Now().UnixNano())
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

func Stop() {
	if mpvClient == nil {
		return
	}

	_ = mpvClient.Stop()
}

func RestartSong() {
	if mpvClient == nil {
		return
	}

	_ = mpvClient.Seek(-int(mpvClient.Position()))
}

func Back10Seconds() {
	if mpvClient == nil {
		return
	}

	_ = mpvClient.Seek(-10)
}

func Forward10Seconds() {
	if mpvClient == nil {
		return
	}

	_ = mpvClient.Seek(+10)
}

func SeekTo(seconds float64) {
	if mpvClient == nil {
		return
	}
	_ = mpvClient.SetProperty("time-pos", seconds)
}

func VolumeUp() {
	if mpvClient == nil {
		return
	}

	newVolume := ((mpvClient.CurrentVolume() / volumeStep) + 1) * volumeStep
	SetVolume(newVolume)
}

func VolumeDown() {
	if mpvClient == nil {
		return
	}

	newVolume := ((mpvClient.CurrentVolume() - 1) / volumeStep) * volumeStep
	SetVolume(newVolume)
}

func GetVolume() float64 {
	if mpvClient == nil {
		return 0
	}

	vol, _ := mpvClient.GetFloatProperty("volume")
	return vol
}

func SetVolume(volume int) {
	if mpvClient == nil {
		return
	}

	if volume < 0 {
		volume = 0
	} else if volume > 100 {
		volume = 100
	}
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
