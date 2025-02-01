package help

import (
	"github.com/bwmarrin/discordgo"
)

func GetHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "Help",
		Description: "Here are the commands you can use with this bot",
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "/ping",
				Value:  "This command will return latency.",
				Inline: false,
			},
			{
				Name:   "/help",
				Value:  "This command will show the list of commands.",
				Inline: false,
			},
			{
				Name:   "/jtc",
				Value:  "This is the join-to-create command.",
				Inline: false,
			},
			{
				Name:   "/jtc set",
				Value:  "This is for setting the channel to be the join-to-create command (only for admin).",
				Inline: false,
			},
			{
				Name:   "/jtc unset",
				Value:  "This is for unsetting the channel.",
				Inline: false,
			},
		},
	}
	s.ChannelMessageSendEmbed(i.ChannelID, embed)
}
