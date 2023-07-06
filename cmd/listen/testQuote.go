package listen

// Commented test scaffolding for testing pulling individual quotes
//

/*
var testQuoteCommand = &discordgo.ApplicationCommand{
	Name:        "quote",
	Type:        discordgo.ChatApplicationCommand,
	Description: "test quote response for given id",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "id",
			Description: "id to pull",
			Type:        discordgo.ApplicationCommandOptionInteger,
			Required:    false,
		},
	},
}

func (s *Server) quoteTestCommand(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	ctx, cancel := util.ContextFromDiscordInteractionCreate(context.Background(), i, interactionTimeout)
	defer cancel()
	var quoteId int64
	for _, option := range i.ApplicationCommandData().Options {
		if option.Name == "id" {
			quoteId = option.IntValue()
		}
	}
	var quote model.Quote
	var err error
	if quoteId == 0 {
		quote, err = s.app.Quotes.GetQuoteSql(ctx)
	} else {
		quote, err = s.app.Quotes.GetQuoteById(ctx, int(quoteId))
	}
	if err != nil {
		slog.ErrorCtx(ctx, "failed to get quote from model: "+err.Error(), "id", quoteId)
		return
	}
	switch quote.Type {
	case model.TextQuoteType:
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: quote.Data,
			},
		})
	case model.FileQuoteType:
		file, err := os.Open(quote.Data)
		if err != nil {
			break
		}
		defer file.Close()
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Files: []*discordgo.File{{
					Name:        filepath.Base(quote.Data),
					ContentType: strings.Replace(".", "", filepath.Ext(quote.Data), 1),
					Reader:      file,
				}},
			},
		})
	default:
		err = model.InvalidQuoteTypeError(quote.Type)
	}
	if err != nil {
		slog.ErrorCtx(ctx, "failed to respond to interaction: "+err.Error(), "quote", quote)
	}

}

*/
