package bot

import (
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"

	"git.declansoper.xyz/japanbot/japanbot-go/bot/config"
	"git.declansoper.xyz/japanbot/japanbot-go/bot/dictionary"
)

// JapanBot is a Discord bot with Japanese parsing abilities
type JapanBot struct {
	dictionary    *dictionary.Dictionary
	configuration *config.BotConfiguration
	session       *discordgo.Session
	handlers      HandlerMap
}

// Start starts the JapanBot instance
func (b *JapanBot) Start() error {
	discord, err := discordgo.New("Bot " + b.configuration.APIToken)
	if err != nil {
		return err
	}

	b.session = discord
	discord.AddHandler(b.onMessageCreate)
	if err = discord.Open(); err != nil {
		return err
	}

	return nil
}

// Stop will stop JapanBot
func (b *JapanBot) Stop() error {
	return b.session.Close()
}

func (b *JapanBot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	args := strings.Split(m.Content, " ")
	command := args[0]
	if len(command) > 4 && strings.Contains(command, "!") {
		keyword := command[4:]
		handler := b.handlers[keyword]
		if handler != nil {
			go handler(args, s, m.Message)
		}
	}
}

// New creates a new instance of JapanBot using a given config
func New(config *config.BotConfiguration) (*JapanBot, error) {
	r, err := os.Open(config.JMdictFile)
	if err != nil {
		return nil, err
	}

	d, err := dictionary.Load(r)
	if err != nil {
		return nil, err
	}

	b := &JapanBot{
		dictionary:    d,
		configuration: config,
	}
	b.handlers = b.createHandlerMap()
	return b, nil
}
