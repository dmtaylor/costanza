package listen

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/dmtaylor/costanza/internal/roller"
)

// dispatchRollCommands Main entrypoint into handling roll commands. Reads the first word of the message content
// and calls the appropriate method for performing a roll. Update this to add additional message prefixes for additional
// roll types.
func (s *Server) dispatchRollCommands(sess *discordgo.Session, m *discordgo.MessageCreate) {
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
