package ping

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func GetPing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	latency := s.HeartbeatLatency().Milliseconds()

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Pong! Latency: %dms", latency),
		},
	}

	if err := s.InteractionRespond(i.Interaction, response); err != nil {
		fmt.Println("Error responding to /ping command:", err)
	}
}
