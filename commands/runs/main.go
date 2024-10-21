package run

import (
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
