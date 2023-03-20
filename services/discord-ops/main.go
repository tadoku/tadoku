package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppID    string `validate:"required" envconfig:"app_id"`
	BotToken string `validate:"required" envconfig:"bot_token"`
	GuildID  string `validate:"required" envconfig:"guild_id"`
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "psqlbackup",
			Description: "Command for demonstrating options",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "immersion",
					Description: "Immersion PostgreSQL database",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "content",
					Description: "Content PostgreSQL database",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "kratos",
					Description: "Kratos PostgreSQL database",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"psqlbackup": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			content := ""

			switch options[0].Name {
			case "immersion":
				content = "kratos backup"
			case "content":
				content = "content backup"
			case "kratos":
				content = "kratos backup"
			default:
				content = "Oops, something went wrong."
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
		},
	}
)

func main() {
	cfg := Config{}
	envconfig.Process("OPS", &cfg)

	validate := validator.New()
	err := validate.Struct(cfg)
	if err != nil {
		panic(fmt.Errorf("could not configure bot: %w", err))
	}

	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		panic(fmt.Errorf("could not configure bot: %w", err))
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	cmdIDs := make(map[string]string, len(commands))

	for _, cmd := range commands {
		rcmd, err := session.ApplicationCommandCreate(cfg.AppID, cfg.GuildID, cmd)
		if err != nil {
			log.Fatalf("Cannot create slash command %q: %v", cmd.Name, err)
		}

		cmdIDs[rcmd.ID] = rcmd.Name

	}

	err = session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer session.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")

	for id, name := range cmdIDs {
		err := session.ApplicationCommandDelete(cfg.AppID, cfg.GuildID, id)
		if err != nil {
			log.Fatalf("Cannot delete slash command %q: %v", name, err)
		}
	}
}
