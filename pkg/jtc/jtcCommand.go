package jtcCommand

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

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

func GetJTC(db *sql.DB, s *discordgo.Session, i *discordgo.InteractionCreate) {

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
	messageSet := messageOptionsSet.StringValue()
	messageID := messageOptionsID.StringValue()

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
