package integration

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/babycommando/rich-go/client"
	"github.com/babycommando/rich-go/ipc"
)

type DiscordInstance struct {
	Connected bool
}

func InitDiscord() *DiscordInstance {
	err := client.Login("1461114330790494334")
	if err != nil {
		log.Printf("[Discord] Could not connect: %v", err)
		return nil
	}
	return &DiscordInstance{Connected: true}
}

func (ins *DiscordInstance) Close() {
	client.Logout()
}

func (ins *DiscordInstance) UpdateActivity(meta Metadata) {
	now := time.Now()

	start := now.Add(-time.Duration(meta.Position) * time.Second)
	end := start.Add(time.Duration(meta.Duration) * time.Second)

	err := client.SetActivity(client.Activity{
		Details:    meta.Title,
		State:      meta.Artist + " - " + meta.Album,
		LargeImage: meta.ImageURL,
		Type:       2, // Listening
		Timestamps: &client.Timestamps{
			Start: &start,
			End:   &end,
		},
	})
	if err != nil {
		log.Printf("[Discord] Update error: %v", err)
	}
}

func (ins *DiscordInstance) StopActivity() {
	if ins == nil || !ins.Connected {
		return
	}

	payload := struct {
		Cmd  string `json:"cmd"`
		Args struct {
			PID int `json:"pid"`
		} `json:"args"`
		Nonce string `json:"nonce"`
	}{
		Cmd:   "SET_ACTIVITY",
		Nonce: "subtui-clear-activity",
	}
	payload.Args.PID = os.Getpid()

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Discord] Could not clear activity: %v", err)
		return
	}

	_ = ipc.Send(1, string(data))
}
