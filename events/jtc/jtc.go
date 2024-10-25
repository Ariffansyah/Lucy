package eventsjtc

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
)

var activeChannels = make(map[string]string)
var channelMembers = make(map[string]map[string]bool)
var activeChannel = make(map[string]bool)

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

func JoinToCreate(db *sql.DB, s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	// Check if user is moving between channels
	if vs.BeforeUpdate != nil && vs.BeforeUpdate.ChannelID == vs.ChannelID {
		// If ChannelID didn't change, it's just a mute/unmute or deafen/undeafen action.
		return
	}

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

	// Handle channel leave logic if BeforeUpdate.ChannelID exists
	if vs.BeforeUpdate != nil && vs.BeforeUpdate.ChannelID != "" {
		previousChannelID := vs.BeforeUpdate.ChannelID
		if channelMembers[previousChannelID] != nil {
			delete(channelMembers[previousChannelID], vs.UserID)
			memberCount := len(channelMembers[previousChannelID])

			if memberCount == 0 && activeChannel[previousChannelID] {
				log.Printf("No members in the dynamically created channel %s, deleting it", previousChannelID)
				_, err := s.ChannelDelete(previousChannelID)
				if err != nil {
					fmt.Printf("Error deleting channel %s: %v\n", previousChannelID, err)
					return
				}
				delete(channelMembers, previousChannelID)
				delete(activeChannels, previousChannelID)
			}
		}
	}

	// Check if the user has joined one of the monitored channels
	for _, channelID := range channelIDs {
		if vs.ChannelID == channelID {
			fmt.Printf("User %s has joined the monitored voice channel %s\n", vs.UserID, channelID)

			// Get parent channel ID
			CategoryID, err := s.Channel(channelID)
			if err != nil {
				fmt.Println("Error fetching channel category:", err)
				return
			}

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

			err = s.GuildMemberMove(vs.GuildID, vs.UserID, &newChannel.ID)
			if err != nil {
				fmt.Printf("Error moving user %s to the new channel: %v\n", vs.UserID, err)
				return
			}

			activeChannel[newChannel.ID] = true
			fmt.Printf("User %s moved to their own channel %s\n", vs.UserID, newChannel.ID)
			return
		}
	}

	// Handle user joining a channel
	if vs.ChannelID != "" {
		if channelMembers[vs.ChannelID] == nil {
			channelMembers[vs.ChannelID] = make(map[string]bool)
		}

		channelMembers[vs.ChannelID][vs.UserID] = true
		memberCount := len(channelMembers[vs.ChannelID])

		log.Printf("User %s joined channel %s", vs.UserID, vs.ChannelID)
		log.Printf("Channel %s now has %d member(s)", vs.ChannelID, memberCount)
	} else if vs.ChannelID == "" {
		for channelID, members := range channelMembers {
			if members[vs.UserID] {
				delete(channelMembers[channelID], vs.UserID)
				memberCount := len(channelMembers[channelID])

				log.Printf("User %s left channel %s", vs.UserID, channelID)
				log.Printf("Channel %s now has %d member(s)", channelID, memberCount)

				if memberCount == 0 && activeChannel[channelID] {
					log.Printf("No members in the dynamically created channel %s, deleting it", channelID)
					_, err := s.ChannelDelete(channelID)
					if err != nil {
						fmt.Printf("Error deleting channel %s: %v\n", channelID, err)
						return
					}
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
