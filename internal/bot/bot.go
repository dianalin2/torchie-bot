package bot

import (
	"flag"
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"math/rand"
)

var dg *discordgo.Session

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "coinflip",
		Description: "50/50",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "choice",
				Description: "Heads or Tails",
				Required:    false,
			},
		},
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"coinflip": coinflip,
}

var (
	guildId = flag.String("g", "", "Guild ID")
	token = flag.String("t", "", "Bot Token")
)

func init() {
	flag.Parse()

	var err error
	dg, err = discordgo.New("Bot " + *token)

	if err != nil {
		fmt.Println("error creating Discord session,", err)
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				handler(s, i)
			}
		}
	})
}

func Run() error {

	err := dg.Open()

	if err != nil {
		fmt.Println("error opening connection,", err)
		return err
	}

	for _, command := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, *guildId, command)
		if err != nil {
			fmt.Println("error creating command,", err)
		}
	}
	
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()

	return nil
}

func coinflip(s *discordgo.Session, i *discordgo.InteractionCreate) {

	options := i.ApplicationCommandData().Options

	side := rand.Intn(2) == 0

	if options != nil && len(options) > 0 {
		var choice string
		if side {
			choice = "It would be very wise to " + options[0].StringValue() + "."
		} else {
			choice = "To " + options[0].StringValue() + " would be a bad idea."
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: choice,
			},
		})
	} else {
		var choice string
		if side {
			choice = "Logic says yes."
		} else {
			choice = "Logic says no."
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: choice,
			},
		})
	}
}
