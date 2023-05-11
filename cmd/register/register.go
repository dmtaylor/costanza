package register

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/cobra"

	"github.com/dmtaylor/costanza/cmd/listen"
	"github.com/dmtaylor/costanza/config"
)

var Cmd = &cobra.Command{
	Use:   "register",
	Short: "register discord commands",
	Long:  "Register all application commands for discord app",
	RunE:  registerCommands,
}

func registerCommands(_ *cobra.Command, _ []string) error {
	err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	log.Printf("starting command registration")
	sess, err := discordgo.New("Bot " + config.GlobalConfig.Discord.Token)
	if err != nil {
		return fmt.Errorf("failed to get discord session: %w", err)
	}
	err = sess.Open()
	if err != nil {
		return fmt.Errorf("failed to open discord connection: %w", err)
	}
	var closeErr error = nil
	defer func() {
		closeErr = sess.Close()
	}()

	for _, command := range listen.Commands {
		log.Printf("registering %s", command.Name)
		_, err := sess.ApplicationCommandCreate(sess.State.User.ID, "", command)
		if err != nil {
			return fmt.Errorf("failed to register command %s: %w", command.Name, err)
		}
	}

	log.Printf("successfully registered all commands")

	return closeErr
}
