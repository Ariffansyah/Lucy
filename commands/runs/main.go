package run

import (
	"Lucy/pkg/help"
	"Lucy/pkg/jtc"
	"Lucy/pkg/ping"

	"database/sql"
	"github.com/bwmarrin/discordgo"
)

func RunPing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ping.GetPing(s, i)
}

func RunJTC(db *sql.DB, s *discordgo.Session, i *discordgo.InteractionCreate) {
	jtcCommand.GetJTC(db, s, i)
}

func RunHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	help.GetHelp(s, i)
}
