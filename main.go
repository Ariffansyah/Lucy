package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	sqlite3 "github.com/mattn/go-sqlite3"
)

var activeChannels = make(map[string]string)

func CommandHandler(db *sql.DB, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Ensure we are working with the right type
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "ping":
			pingCommand(s, i)
		case "jtc":
			jtcCommand(db, s, i)
		}
	}

	log.Printf("Command Executed")
}

func pingCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

type ChannelID struct {
	channe_id string
}

func getUsers(db *sql.DB) ([]ChannelID, error) {
	// Execute the query
	rows, err := db.Query("SELECT id, name, email FROM users") // Replace 'users' with your table name
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []ChannelID
	for rows.Next() {
		var user ChannelID
		// Scan the result into the User struct
		if err := rows.Scan(&user.channe_id); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func getChannelByID(s *discordgo.Session, channelID string) (*discordgo.Channel, error) {
	channel, err := s.Channel(channelID)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func jtcCommand(db *sql.DB, s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandData := i.ApplicationCommandData()

	// Check if there are enough options
	if len(commandData.Options) < 2 {
		// Respond with an error message if options are missing
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide both options.",
			},
		})
		return
	}

	// Extract the options safely
	messageOptionsSet := commandData.Options[0]
	messageOptionsID := commandData.Options[1]

	// Safely access the string values from the options
	messageSet := messageOptionsSet.StringValue() // Use StringValue() for the first option
	messageID := messageOptionsID.StringValue()   // Corrected to access the first option Ensure the database connection is closed

	// Create the channel table if it doesn't exist

	switch messageSet {
	case "set":
		channel, err := getChannelByID(s, messageID)

		if err != nil {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Channel ID %s does not exist.", messageID),
				},
			}
			if err := s.InteractionRespond(i.Interaction, response); err != nil {
				fmt.Println("Error responding to /jtc command:", err)
			}
			return
		}

		if channel.Type != discordgo.ChannelTypeGuildVoice {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "This command can only be used in voice channels.",
				},
			}
			if err := s.InteractionRespond(i.Interaction, response); err != nil {
				fmt.Println("Error responding to /jtc command:", err)
			}
			return
		}

		saveChannelID(db, messageID)
		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Channel ID %s saved to the database.", messageID),
			},
		}
		if err := s.InteractionRespond(i.Interaction, response); err != nil {
			fmt.Println("Error responding to /jtc command:", err)
		}
	case "unset":
		_, err := getChannelByID(s, messageID)

		if err != nil {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Channel ID %s does not exist.", messageID),
				},
			}
			if err := s.InteractionRespond(i.Interaction, response); err != nil {
				fmt.Println("Error responding to /jtc command:", err)
			}
		}

		deleteChannelID(db, messageID)
		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Channel ID %s deleted from the database.", messageID),
			},
		}
		if err := s.InteractionRespond(i.Interaction, response); err != nil {
			fmt.Println("Error responding to /jtc command:", err)
		}
	default:
		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid option. Please use either 'set' or 'unset'.",
			},
		}
		if err := s.InteractionRespond(i.Interaction, response); err != nil {
			fmt.Println("Error responding to /jtc command:", err)
		}
	}
}

func JoinToCreate(db *sql.DB, s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	rows, err := db.Query("SELECT channel_id FROM channel")
	if err != nil {
		fmt.Println("Error querying database:", err)
		return
	}
	defer rows.Close()

	// Store channel IDs
	var channelIDs []string
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			fmt.Println("Error scanning row:", err)
			return
		}
		channelIDs = append(channelIDs, channelID)
	}

	// Check if the user has joined one of the monitored channels
	for _, channelID := range channelIDs {
		if vs.ChannelID == channelID {
			// User has joined the target channel
			fmt.Printf("User %s has joined the monitored voice channel %s\n", vs.UserID, channelID)

			// Get parent channel ID
			CategoryID, err := s.Channel(channelID)

			// Create a new voice channel and move the user to it
			newChannel, err := s.GuildChannelCreateComplex(vs.GuildID, discordgo.GuildChannelCreateData{
				Name:     fmt.Sprintf("%s's Channel", vs.Member.User.Username),
				Type:     discordgo.ChannelTypeGuildVoice,
				ParentID: CategoryID.ParentID,
			})
			if err != nil {
				fmt.Println("Error creating channel:", err)
				return
			}

			// Move the user to the newly created channel
			err = s.GuildMemberMove(vs.GuildID, vs.UserID, &newChannel.ID)
			if err != nil {
				fmt.Printf("Error moving user %s to the new channel: %v\n", vs.UserID, err)
				return
			}

			// Store the created channel ID so it can be deleted later
			activeChannels[vs.UserID] = newChannel.ID
			fmt.Printf("User %s moved to their own channel %s\n", vs.UserID, newChannel.ID)

			return
		}
	}

	// Check if the user has left a newly created channel
	if channelID, ok := activeChannels[vs.UserID]; ok {
		if vs.ChannelID != channelID {
			// delete the channel
			_, err := s.ChannelDelete(channelID)
			if err != nil {
				fmt.Printf("Error deleting channel %s: %v\n", channelID, err)
				return
			}
			fmt.Printf("Channel %s deleted\n", channelID)
		}
	}
}

func RegisterCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Replies with a ping!",
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
		JoinToCreate(db, s, vs)
	})

	if err := dg.Open(); err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	RegisterCommands(dg)

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	select {}
}

func saveChannelID(db *sql.DB, channelID string) {
	// Insert the channel ID into the database
	stmt, err := db.Prepare("INSERT INTO channel(channel_id) VALUES(?)")
	if err != nil {
		fmt.Println("Error preparing statement:", err)
		return
	}
	defer stmt.Close()

	// Execute the statement
	_, err = stmt.Exec(channelID)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			fmt.Printf("Channel ID %s already exists in the database.\n", channelID)
		} else {
			fmt.Println("Error executing statement:", err)
		}
		return
	}
	fmt.Printf("Channel ID %s saved to the database.\n", channelID)
}

func deleteChannelID(db *sql.DB, channelID string) {
	// Delete the channel ID from the database
	stmt, err := db.Prepare("DELETE FROM channel WHERE channel_id = ?")
	if err != nil {
		fmt.Println("Error preparing statement:", err)
		return
	}
	defer stmt.Close()

	// Execute the statement
	_, err = stmt.Exec(channelID)
	if err != nil {
		fmt.Println("Error executing statement:", err)
		return
	}
	fmt.Printf("Channel ID %s deleted from the database.\n", channelID)
}
