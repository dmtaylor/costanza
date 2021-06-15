package server

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/dmtaylor/costanza/internal/quotes"
	"github.com/pkg/errors"
)

type Server struct {
	quotes *quotes.QuoteEngine
	//roller Roller TODO
}

func New() (*Server, error) {
	quoteEngine, err := quotes.NewQuoteEngine()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build quote engine")
	}
	return &Server{
		quotes: quoteEngine,
	}, nil
}

func (s *Server) MessageCreate(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	for _, mentionedUser := range m.Mentions {
		if mentionedUser.ID == sess.State.User.ID {
			log.Printf("got message\n")
			_, err := sess.ChannelMessageSend(m.ChannelID, s.quotes.GetQuote())
			if err != nil {
				log.Printf("error sending message: %s\n", err)
			}
			return
		}
	}

}
