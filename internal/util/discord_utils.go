package util

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
)

// UserIsServerAdmin is a helper function for if a given member can manage the guild
func UserIsServerAdmin(member *discordgo.Member) bool {
	return member.Permissions&discordgo.PermissionManageGuild == discordgo.PermissionManageGuild
}

// MessageExcluded helper function to determine if a message shouldn't be handled
func MessageExcluded(sess *discordgo.Session, m *discordgo.MessageCreate) bool {
	return m.Author.Bot || m.Author.ID == sess.State.User.ID
}

func MustSnowflakeToInt(snowflake string) uint64 {
	i, err := strconv.ParseUint(snowflake, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}
