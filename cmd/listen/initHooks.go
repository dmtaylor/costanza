package listen

import "github.com/bwmarrin/discordgo"

func (s *Server) StartInitative(sess *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO implement entrypoint for starting initiative
}

func (s *Server) NextInitiative(sess *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO implement entrypoint for advancing initiative
}

func (s *Server) EndInitiative(sess *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO implement entrypoint for ending init order by deleting session
}
