package util

import "github.com/bwmarrin/discordgo"

const maxMessageSize = 2000

func ChannelMessageSendChunked(sess *discordgo.Session, channelId string, content string) (*discordgo.Message, error) {

	// TODO implement this
	return nil, nil
}

func ChannelMessageSendReplyChunked(sess *discordgo.Session, channelId string, content string, reference *discordgo.MessageReference) (*discordgo.Message, error) {

	// TODO implement this
	return nil, nil

}
