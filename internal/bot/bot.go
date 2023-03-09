package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"math/rand"
	"regexp"
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
				Description: "Choice to make",
				Required:    false,
			},
		},
	},
	{
		Name: "choice",
		Description: "Make a choice",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "choices",
				Description: "Choices to choose from",
				Required:    true,
			},
		},
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"coinflip": coinflip,
	"choice": choice,
}

var guildId string
var token string
var optionRegex = regexp.MustCompile(`"([^"]|\\")+"|(\\"|[^" \n\t])+`)

func init() {
	godotenv.Load()

	guildId = os.Getenv("GUILD_ID")
	token = os.Getenv("BOT_TOKEN")

	var err error
	dg, err = discordgo.New("Bot " + token)

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

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, command := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, guildId, command)
		if err != nil {
			fmt.Println("error creating command,", err)
		} else {
			registeredCommands[i] = command
		}
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	for _, command := range registeredCommands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, command.GuildID, command.ID)
		if err != nil {
			fmt.Println("error deleting command,", err)
		}
	}

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

func choice(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	choices := optionRegex.FindAllString(options[0].StringValue(), -1)
	choice := rand.Intn(len(choices))
	content := choices[choice]

	if content[0] == '"' {
		content = content[1:len(content)-1]
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}
