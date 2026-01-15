package integration

import (
	"log"
	"time"

	"github.com/babycommando/rich-go/client"
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

func (ins *DiscordInstance) UpdateActivity(meta Metadata) {
	now := time.Now()

	start := now.Add(-time.Duration(meta.Position) * time.Second)
	end := start.Add(time.Duration(meta.Duration) * time.Second)

	err := client.SetActivity(client.Activity{
		Details:    meta.Title,
		State:      meta.Artist + " - " + meta.Album,
		LargeImage: meta.ImageURL,
		LargeText:  meta.Album,
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
