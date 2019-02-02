package bot

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/hakasec/japanbot-go/bot/config"
	"github.com/hakasec/japanbot-go/bot/database"
	"github.com/hakasec/japanbot-go/bot/database/models"
	"github.com/hakasec/japanbot-go/bot/database/set"
	"github.com/hakasec/japanbot-go/bot/dictionary"
)

// JapanBot is a Discord bot with Japanese parsing abilities
type JapanBot struct {
	dictionary    *dictionary.Dictionary
	configuration *config.BotConfiguration
	session       *discordgo.Session
	db            *database.DBConnection
	handlers      HandlerMap

	channels *set.DBSet
	cards    *set.DBSet

	analyseRequests map[string][]string
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
		keywords := strings.Split(command[4:], "!")
		handler := b.handlers[keywords[0]]
		if handler != nil {
			go handler(args, s, m.Message)
			return
		}
	}

	var channel models.Channel
	b.channels.Get(
		map[string]interface{}{
			"ChannelID": m.ChannelID,
		},
		&channel,
	)

	postCard := rand.Intn(10) == 0
	if channel.CardMode != 0 && postCard {
		card := b.generateCard(m.ChannelID)
		if err := b.cards.Add(card); err != nil {
			fmt.Printf("Error adding card: %s\n", err.Error())
		} else {
			s.ChannelMessageSend(
				m.ChannelID,
				fmt.Sprintf("```Card:\nPhrase: %s\n```", card.Phrase),
			)
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

	db, err := database.OpenFromConfig(&config.DBConfig)
	if err != nil {
		return nil, err
	}

	channelSet := set.New("channels", reflect.TypeOf(models.Channel{}), db)
	err = channelSet.CreateTable()
	if err != nil {
		return nil, err
	}
	cardSet := set.New("cards", reflect.TypeOf(models.Card{}), db)
	err = cardSet.CreateTable()
	if err != nil {
		return nil, err
	}

	b := &JapanBot{
		dictionary:    d,
		db:            db,
		configuration: config,

		channels: channelSet,
		cards:    cardSet,

		analyseRequests: make(map[string][]string),
	}
	b.handlers = b.createHandlerMap()
	return b, nil
}
