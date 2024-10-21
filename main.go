package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

var activeChannels = make(map[string]string)
var channelMembers = make(map[string]map[string]bool)
var activeChannel = make(map[string]bool)

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

	// Fetch the member to check their permissions
	_, err := s.GuildMember(i.GuildID, i.Member.User.ID)
	if err != nil {
		log.Printf("Error fetching member details: %v", err)
		return
	}

	// Get the permissions of the member in the current guild and check for administrator rights
	userPermissions, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil {
		log.Printf("Error fetching permissions: %v", err)
		return
	}

	if userPermissions&discordgo.PermissionAdministrator == 0 {
		// User is not an administrator
		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not have permission to use this command. Admins only.",
			},
		}
		if err := s.InteractionRespond(i.Interaction, response); err != nil {
			fmt.Println("Error responding to /jtc command:", err)
		}
		return
	}

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
	messageID := messageOptionsID.StringValue()   // Corrected to access the first option

	// Perform the action based on the messageSet value ("set" or "unset")
	switch messageSet {
	case "set":
		// Get channel details by ID
		channel, err := s.Channel(messageID)
		if err != nil {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Channel ID %s does not exist in this server.", messageID),
				},
			}
			if err := s.InteractionRespond(i.Interaction, response); err != nil {
				fmt.Println("Error responding to /jtc command:", err)
			}
			return
		}

		// Ensure the channel is from the same guild as the one where the command is issued
		if channel.GuildID != i.GuildID {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You can only set channels from the same server.",
				},
			}
			if err := s.InteractionRespond(i.Interaction, response); err != nil {
				fmt.Println("Error responding to /jtc command:", err)
			}
			return
		}

		// Ensure the channel is a voice channel
		if channel.Type != discordgo.ChannelTypeGuildVoice {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "This command can only be used for voice channels.",
				},
			}
			if err := s.InteractionRespond(i.Interaction, response); err != nil {
				fmt.Println("Error responding to /jtc command:", err)
			}
			return
		}

		// Save the channel ID to the database
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
		// Check if the channel exists in the same guild
		channel, err := s.Channel(messageID)
		if err != nil || channel.GuildID != i.GuildID {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Channel ID %s does not exist in this server.", messageID),
				},
			}
			if err := s.InteractionRespond(i.Interaction, response); err != nil {
				fmt.Println("Error responding to /jtc command:", err)
			}
			return
		}

		// Ensure the channel is from the same guild as the one where the command is issued
		if channel.GuildID != i.GuildID {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You can only set channels from the same server.",
				},
			}
			if err := s.InteractionRespond(i.Interaction, response); err != nil {
				fmt.Println("Error responding to /jtc command:", err)
			}
			return
		}

		// Delete the channel ID from the database
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
		// Handle invalid options
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

	// Detect when a user moves between channels
	// If `vs.BeforeUpdate.ChannelID` is not empty, the user has moved from another channel
	if vs.BeforeUpdate != nil && vs.BeforeUpdate.ChannelID != "" {
		previousChannelID := vs.BeforeUpdate.ChannelID
		// Remove the user from the previous channel's member list
		if channelMembers[previousChannelID] != nil {
			delete(channelMembers[previousChannelID], vs.UserID)
			memberCount := len(channelMembers[previousChannelID])

			// Check if the previous channel is dynamically created and empty
			if memberCount == 0 && activeChannel[previousChannelID] {
				log.Printf("No members in the dynamically created channel %s, deleting it", previousChannelID)
				_, err := s.ChannelDelete(previousChannelID)
				if err != nil {
					fmt.Printf("Error deleting channel %s: %v\n", previousChannelID, err)
					return
				}
				fmt.Printf("Dynamically created channel %s deleted\n", previousChannelID)
				delete(channelMembers, previousChannelID)
				delete(activeChannels, previousChannelID)
			}
		}
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

			// Track the dynamically created channel
			activeChannel[newChannel.ID] = true
			fmt.Printf("User %s moved to their own channel %s\n", vs.UserID, newChannel.ID)

			return
		}
	}

	// Handle user joining a channel
	if vs.ChannelID != "" {
		// Ensure the channel has an entry in the map
		if channelMembers[vs.ChannelID] == nil {
			channelMembers[vs.ChannelID] = make(map[string]bool)
		}

		// Add the user to the channel's member list
		channelMembers[vs.ChannelID][vs.UserID] = true
		memberCount := len(channelMembers[vs.ChannelID])

		log.Printf("User %s joined channel %s", vs.UserID, vs.ChannelID)
		log.Printf("Channel %s now has %d member(s)", vs.ChannelID, memberCount)
	} else if vs.ChannelID == "" { // User leaves the channel completely
		for channelID, members := range channelMembers {
			// Check if the user was in this channel
			if members[vs.UserID] {
				// Remove the user from the channel's member list
				delete(channelMembers[channelID], vs.UserID)
				memberCount := len(channelMembers[channelID])

				log.Printf("User %s left channel %s", vs.UserID, channelID)
				log.Printf("Channel %s now has %d member(s)", channelID, memberCount)

				// If no members remain and the channel was dynamically created, delete the channel
				if memberCount == 0 && activeChannel[channelID] {
					log.Printf("No members in the dynamically created channel %s, deleting it", channelID)
					_, err := s.ChannelDelete(channelID)
					if err != nil {
						fmt.Printf("Error deleting channel %s: %v\n", channelID, err)
						return
					}
					fmt.Printf("Dynamically created channel %s deleted\n", channelID)
					delete(channelMembers, channelID)
					delete(activeChannels, channelID)
				}
				break
			}
		}
	}
}

// Helper function to check if the channel was dynamically created
func isDynamicallyCreated(channelID string) bool {
	for _, createdChannelID := range activeChannels {
		if createdChannelID == channelID {
			return true
		}
	}
	return false
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
