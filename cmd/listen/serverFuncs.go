package listen

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/roller"
)

var startLateHours, endLateHours time.Time

const helpMessage string = `
Costanza commands:
` +
	"```" + `
!chelp:   this message.
!roll:    parse text as d-notation and evaluate expression.
!srroll:  parse text as d-notation, evaluate, and use result for Shadowrun roll.
!wodroll: parse text as d-notation, evaluate, and use result for World of Darkness roll.
          Can be modified with '8again', '9again' and 'chance'. Rolls of < 1 dice are done as chance rolls.
!dhtest:  parse text as d-notation, evaluate, and use result for FF Warhammer 40k RPG roll (over-under on 1d100).
` +
	"```"

// Help handler function for help messages
func (s *Server) Help(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	words := strings.Fields(m.Message.Content)
	if len(words) < 1 {
		return
	}
	if words[0] == "!chelp" {
		_, err := sess.ChannelMessageSend(
			m.ChannelID,
			helpMessage,
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
	}
}

// EchoQuote handler function for sending George Costanza quotes
func (s *Server) EchoQuote(sess *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()
	if m.Author.ID == sess.State.User.ID {
		return
	}

	for _, mentionedUser := range m.Mentions {
		if mentionedUser.ID == sess.State.User.ID {
			s.sendQuote(ctx, sess, m)
			return
		}
	}
}

func (s *Server) sendQuote(ctx context.Context, sess *discordgo.Session, m *discordgo.MessageCreate) {
	quote, err := s.app.Quotes.GetQuoteSql(ctx)
	if err != nil {
		log.Printf("failed to get quote: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to get a quote. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	_, err = sess.ChannelMessageSendReply(m.ChannelID, quote, m.Reference())
	if err != nil {
		log.Printf("error sending message: %s\n", err)
	}
}

func (s *Server) EchoInsomniac(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	if s.isInsomniacUser(m.Author, m.Member) && isAfterHours() {
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf("%s All right. That's enough for today. You're tired. Get some sleep. I'll see you first thing in the morning.",
				m.Author.Mention()),
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}

}

// DispatchRollCommands Main entrypoint into handling roll commands. Reads the first word of the message content
// and calls the appropriate method for performing a roll. Update this to add additional message prefixes for additional
// roll types.
func (s *Server) DispatchRollCommands(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}
	command := strings.Fields(m.Message.Content)
	if len(command) < 1 {
		return
	}

	switch command[0] {
	case "!roll":
		s.doDNotationRoll(sess, m, strings.Join(command[1:], " "))
	case "!srroll":
		s.doShadowrunRoll(sess, m, strings.Join(command[1:], " "))
	case "!wodroll":
		s.doWodRoll(sess, m, command[1:])
	case "!dhtest":
		s.doDHTestRoll(sess, m, strings.Join(command[1:], " "))
	}
}

func (s *Server) doDNotationRoll(sess *discordgo.Session, m *discordgo.MessageCreate, rollStr string) {
	result, err := s.app.DNotationParser.DoParse(rollStr)
	if err != nil {
		log.Printf("error parsing string: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf("I was unable to understand your roll \"%s\". Why must there always be a problem?", rollStr),
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	response := fmt.Sprintf("%s = %d", result.StrValue, result.Value)
	_, err = sess.ChannelMessageSendReply(
		m.ChannelID,
		response,
		m.Reference(),
	)
	if err != nil {
		log.Printf("error sending message: %s\n", err)
	}
}

func (s *Server) doShadowrunRoll(sess *discordgo.Session, m *discordgo.MessageCreate, rollStr string) {
	rollCount, err := s.app.DNotationParser.DoParse(rollStr)
	if err != nil {
		log.Printf("error parsing string: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf("I was unable to understand your roll \"%s\". Why must there always be a problem?", rollStr),
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	params := roller.GetSrParams()
	result, err := s.app.ThresholdRoller.DoThresholdRoll(rollCount.Value, roller.SrDieSides, params)
	if err != nil {
		log.Printf("failed to do threshold roll: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to perform your roll. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	resultStr, err := result.Repr()
	if err != nil {
		log.Printf("failed to get string repr: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to say what your roll looks like. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	response := fmt.Sprintf("%s = %d", resultStr, result.Value())
	_, err = sess.ChannelMessageSendReply(
		m.ChannelID,
		response,
		m.Reference(),
	)
	if err != nil {
		log.Printf("error sending message: %s\n", err)
	}
	switch roller.GetGlitchStatus(result) {
	case roller.SrGlitch:
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"You glitched! I can't believe this! What was wrong with it? What didn't you like about it?",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
	case roller.SrCritGlitch:
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"You critically glitched! I don't want hope. Hope is killing me. My dream is to become hopeless. When you're hopeless, you don't care, and when you don't care, that indifference makes you attractive.",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
	}
}

func (s *Server) doWodRoll(sess *discordgo.Session, m *discordgo.MessageCreate, tokens []string) {
	params, isChance, rollStr, err := roller.GetWodRollParams(tokens)
	if err != nil {
		log.Printf("failed getting wod params: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to get the params for your roll. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	if isChance {
		s.doWodChanceRoll(sess, m, params)
		return
	}
	rollCount, err := s.app.DNotationParser.DoParse(rollStr)
	if err != nil {
		log.Printf("failed getting number or dice to roll: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to figure out the number of dice to roll. Life can be so confusing. I..I'm searching for answers, anywhere.",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	if rollCount.Value < 1 {
		s.doWodChanceRoll(sess, m, params)
		return
	}
	roll, err := s.app.ThresholdRoller.DoThresholdRoll(rollCount.Value, roller.WodDieSides, params)
	if err != nil {
		log.Printf("failed doing wod threshold roll: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to complete your roll. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	rollResStr, err := roll.Repr()
	if err != nil {
		log.Printf("failed to get result string: %s\n", err)
		return
	}
	response := fmt.Sprintf("%s = %d hits", rollResStr, roll.Value())
	if roll.Value() == 0 {
		response = response + "\nWould you like to critically fail?"
	}
	_, err = sess.ChannelMessageSendReply(
		m.ChannelID,
		response,
		m.Reference(),
	)
	if err != nil {
		log.Printf("error sending message: %s\n", err)
	}
}

func (s *Server) doWodChanceRoll(sess *discordgo.Session, m *discordgo.MessageCreate, params roller.ThresholdParameters) {
	roll, err := s.app.ThresholdRoller.DoThresholdRoll(1, roller.WodDieSides, params)
	if err != nil {
		log.Printf("failed doing wod chance roll: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to complete your chance roll. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	rollResStr, err := roll.Repr()
	if err != nil {
		log.Printf("failed to get string representation: %s\n", err)
		return
	}
	response := fmt.Sprintf("%s = %d hits", rollResStr, roll.Value())
	if roll.Value() == 0 {
		response = response + "\nYou critically failed! Radiating waves of pain."
	}
	_, err = sess.ChannelMessageSendReply(
		m.ChannelID,
		response,
		m.Reference(),
	)
	if err != nil {
		log.Printf("error sending message: %s\n", err)
	}

}

func (s *Server) doDHTestRoll(sess *discordgo.Session, m *discordgo.MessageCreate, rollStr string) {
	threshold, err := s.app.DNotationParser.DoParse(rollStr)
	if err != nil {
		log.Printf("error parsing string: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf("I was unable to understand your roll \"%s\". Why must there always be a problem?", rollStr),
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	roll, err := s.app.DNotationParser.DoParse("1d100")
	if err != nil {
		log.Printf("error doing d100 roll: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to execute the roll. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	var result string
	if roll.Value > threshold.Value {
		degrees := (roll.Value - threshold.Value) / 10
		result = fmt.Sprintf("You Fail with %d degrees", degrees)
	} else {
		degrees := (threshold.Value - roll.Value) / 10
		result = fmt.Sprintf("You Succeed with %d degrees", degrees)
	}
	response := fmt.Sprintf("Rolled %s: %s", roll.StrValue, result)
	_, err = sess.ChannelMessageSendReply(
		m.ChannelID,
		response,
		m.Reference(),
	)
	if err != nil {
		log.Printf("error sending message: %s\n", err)
	}
}

func (s *Server) isInsomniacUser(user *discordgo.User, member *discordgo.Member) bool {
	if user == nil || member == nil {
		return false
	}

	for _, uid := range config.GlobalConfig.Discord.InsomniacIds {
		if user.ID == uid {
			return true
		}
	}

	for _, role := range config.GlobalConfig.Discord.InsomniacRoles {
		for _, userRole := range member.Roles {
			if role == userRole {
				return true
			}
		}
	}
	return false

}

// DailyWinReact performs reaction if it detects a win pattern in the message
func (s *Server) DailyWinReact(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	for _, pattern := range s.dailyWinPatterns {
		if pattern.MatchString(m.Message.Content) {
			err := sess.MessageReactionAdd(m.ChannelID, m.Message.ID, "💯")
			if err != nil {
				log.Printf("error adding reaction: %s\n", err)
			}
			return
		}
	}
}

func (s *Server) LogMessageActivity(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	if m.Author.Bot {
		return
	}

	// Only log stats if channel included in configs
	if _, found := config.GlobalConfig.Discord.ListenChannelSet[m.GuildID]; !found {
		return
	}

	if m.Type == discordgo.MessageTypeDefault || m.Type == discordgo.MessageTypeReply {
		guildId, err := strconv.ParseUint(m.GuildID, 10, 64)
		if err != nil {
			log.Printf("error logging activity: %s\n", err)
			return
		}
		userId, err := strconv.ParseUint(m.Author.ID, 10, 64)
		if err != nil {
			log.Printf("error logging activity: %s\n", err)
			return
		}
		err = s.app.Stats.LogActivity(context.Background(), guildId, userId, m.Timestamp.Format("2006-01"))
		if err != nil {
			log.Printf("error creating activity log: %s\n", err)
		}
	}
}

func isAfterHours() bool {
	var err error
	if startLateHours.IsZero() {
		startLateHours, err = time.Parse(time.Kitchen, "12:30AM")
		if err != nil {
			log.Panicf("Error parsing start date format: %s\n", err)
		}
	}
	if endLateHours.IsZero() {
		endLateHours, err = time.Parse(time.Kitchen, "06:00AM")
		if err != nil {
			log.Panicf("Error parsing end date format: %s\n", err)
		}
	}

	currentTime, err := time.Parse(time.Kitchen, time.Now().Format(time.Kitchen))
	if err != nil {
		log.Printf("failed to parse current time: %s. Failing closed\n", err)
		return false
	}
	return startLateHours.Before(currentTime) && endLateHours.After(currentTime)
}
