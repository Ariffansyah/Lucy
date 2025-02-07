package help

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func GetHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		log.Printf("Error fetching commands: %v", err)
		return
	}

	var fields []*discordgo.MessageEmbedField
	for _, command := range commands {

		if command.Description != "" {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "/" + command.Name,
				Value:  command.Description,
				Inline: false,
			})
		}
	}

	if len(fields) == 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "No Commands Available",
			Value:  "It seems like there are no commands registered for this bot.",
			Inline: false,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Help",
		Description: "Here are the commands you can use with this bot:",
		Color:       0x00ff00, // Green color
		Fields:      fields,
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	if err != nil {
		log.Printf("Error sending interaction response: %v", err)
	}
}
