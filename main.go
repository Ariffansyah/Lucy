package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	cmd "Lucy/commands/runs"
	"Lucy/events/jtc"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func CommandHandler(db *sql.DB, s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "help":
			cmd.RunHelp(s, i)
		case "ping":
			cmd.RunPing(s, i)
		case "jtc":
			cmd.RunJTC(db, s, i)
		}
	}

	log.Printf("Command Executed")
}

func RegisterCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Replies with a ping!",
		},
		{
			Name:        "help",
			Description: "Get help with commands.",
		},
		{
			Name:        "jtc",
			Description: "Join to create!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "subcommand",
					Description: "Set or Unset.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "channelid",
					Description: "Channel ID to set.",
					Required:    true,
				},
			},
		},
	}

	appID := s.State.User.ID // This is your application's ID

	// Loop through and register each command
	for _, cmd := range commands {
		// Register the command globally (empty string for GuildID means global)
		if _, err := s.ApplicationCommandCreate(appID, "", cmd); err != nil {
			fmt.Println("Cannot create command:", cmd.Name, err)
		}
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dg, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	db, err := sql.Open("sqlite3", "./jtc.db")
	if err != nil {
		fmt.Println("error opening database,", err)
		return
	}
	defer db.Close() // Ensure the database connection is closed

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS channel (
		channel_id TEXT PRIMARY KEY
	);`
	if _, err := db.Exec(sqlStmt); err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
		return
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		CommandHandler(db, s, i)
	})

	dg.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		eventsjtc.JoinToCreate(db, s, vs)
	})

	if err := dg.Open(); err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	RegisterCommands(dg)

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	select {}
}
