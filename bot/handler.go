package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hakasec/japanbot-go/bot/database/models"
	"github.com/hakasec/japanbot-go/bot/helpers"
)

// HandlerFunc defines a handler function
// args:	arguments used in command
// s:		discord session
// m:		discord message
type HandlerFunc func(args []string, s *discordgo.Session, m *discordgo.Message)

// HandlerMap is a map of commands to HandlerFuncs
type HandlerMap map[string]HandlerFunc

func (b *JapanBot) createHandlerMap() HandlerMap {
	return HandlerMap{
		"analyze": b.analyse,
		"analyse": b.analyse,
		"answer":  b.answer,
		"help":    b.help,
		"hentai":  b.hentai,
		"enable":  b.enableFeature,
		"disable": b.disableFeature,
	}
}

func (b *JapanBot) analyse(args []string, s *discordgo.Session, m *discordgo.Message) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "You haven't entered a phrase!")
		return
	}

	phrase := strings.Join(args[1:], " ")
	var allGrams []string
	for ngramSize := 1; ngramSize < len(phrase); ngramSize++ {
		tmpNgram := helpers.CreateNgrams(phrase, ngramSize)
		for _, gram := range tmpNgram {
			if !helpers.StringSliceContains(allGrams, gram) {
				allGrams = append(allGrams, gram)
			}
		}
	}

	commandSeg := strings.Split(args[0], "!")
	commandLang := "eng"
	if len(commandSeg) >= 3 {
		commandLang = commandSeg[2]
	}
	hasContent := false
	for _, gram := range allGrams {
		entry, ok := b.dictionary.Index[gram]
		if ok {
			var message strings.Builder
			message.WriteString(fmt.Sprintf("```\n%s:\n", gram))
			hasDefinition := false
			for _, sense := range entry.Senses {
				for _, gloss := range sense.GlossaryItems {
					language := gloss.Language
					if language == "" {
						language = "eng"
					}
					if language == commandLang {
						hasDefinition = true
						message.WriteString(
							fmt.Sprintf("\t%s\n", gloss.Definition),
						)
					}
				}
			}
			message.WriteString("```")
			if hasDefinition {
				hasContent = true
				s.ChannelMessageSend(m.ChannelID, message.String())
			}
		}
	}
	if !hasContent {
		s.ChannelMessageSend(m.ChannelID, "ごめん! No definitions found!")
	}
}

func (b *JapanBot) help(args []string, s *discordgo.Session, m *discordgo.Message) {
	var builder strings.Builder
	builder.WriteString("```\n")

	builder.WriteString(
		`All commands begin with "jpn!"

Available Commands:

- analyse/analyze: Analyse a Japanese sentence. 
  You can add ![language_code] to return results of a chosen changed 
  e.g. jpn!anaylse!spa will return all avaliable Spanish definitions.

- help: This help text, silly!  
`,
	)

	builder.WriteString("```")

	s.ChannelMessageSend(
		m.ChannelID,
		builder.String(),
	)
}

func (b *JapanBot) hentai(args []string, s *discordgo.Session, m *discordgo.Message) {
	s.ChannelMessageSend(m.ChannelID, "CUMMING SOON!")
}

func (b *JapanBot) enableFeature(args []string, s *discordgo.Session, m *discordgo.Message) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "You need to enter the feature you'd like to enable!")
		return
	} else if len(args) > 2 {
		s.ChannelMessageSend(m.ChannelID, "Only one feature at a time please!")
		return
	}

	switch feature := strings.ToLower(args[1]); feature {
	case "card":
		if err := b.changeCardMode(m.ChannelID, 1); err != nil {
			s.ChannelMessageSend(
				m.ChannelID,
				fmt.Sprintf("That failed: %s", err.Error()),
			)
		} else {
			s.ChannelMessageSend(m.ChannelID, "Done :)")
		}
	default:
		s.ChannelMessageSend(m.ChannelID, "That isn't a valid feature!")
	}
}

func (b *JapanBot) disableFeature(args []string, s *discordgo.Session, m *discordgo.Message) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "You need to enter the feature you'd like to disable!")
		return
	} else if len(args) > 2 {
		s.ChannelMessageSend(m.ChannelID, "Only one feature at a time please!")
		return
	}

	switch feature := strings.ToLower(args[1]); feature {
	case "card":
		if err := b.changeCardMode(m.ChannelID, 0); err != nil {
			s.ChannelMessageSend(
				m.ChannelID,
				fmt.Sprintf("That failed: %s", err.Error()),
			)
		} else {
			s.ChannelMessageSend(m.ChannelID, "Done :)")
		}
	default:
		s.ChannelMessageSend(m.ChannelID, "That isn't a valid feature!")
	}
}

func (b *JapanBot) changeCardMode(channelID string, cardMode int) error {
	valueMap := map[string]interface{}{
		"ChannelID": channelID,
	}
	c := &models.Channel{}
	err := b.channels.Get(valueMap, c)
	if err != nil {
		if err == sql.ErrNoRows {
			if cardMode == -1 {
				cardMode = 1
			}
			c.ChannelID = channelID
			c.CardMode = cardMode
			err = b.channels.Add(c)
			return err
		}
		return err
	}

	if cardMode == -1 {
		if c.CardMode == 0 {
			cardMode = 1
		} else {
			cardMode = 0
		}
	}
	c.CardMode = cardMode
	err = b.channels.Update(c)
	return err
}

func (b *JapanBot) generateCard(channelID string) *models.Card {
	rnd := rand.Intn(len(b.dictionary.Entries))
	rndEntry := &b.dictionary.Entries[rnd]

	var phrase string
	if len(rndEntry.KanjiElements) > 0 {
		rnd = rand.Intn(len(rndEntry.KanjiElements))
		phrase = rndEntry.KanjiElements[rnd].Phrase
	} else if len(rndEntry.ReadingElements) > 0 {
		rnd = rand.Intn(len(rndEntry.ReadingElements))
		phrase = rndEntry.ReadingElements[rnd].Phrase
	} else {
		panic(errors.New("Couldn't generate card"))
	}

	return &models.Card{
		ChannelID: channelID,
		Phrase:    phrase,
		EntryID:   rndEntry.EntryID,
		Timestamp: time.Now(),
	}
}

func (b *JapanBot) getLatestCard(channelID string) *models.Card {
	var c models.Card
	err := b.cards.GetDesc(
		map[string]interface{}{
			"ChannelID": channelID,
		},
		"Timestamp",
		&c,
	)
	if err != nil {
		panic(err)
	}
	return &c
}

func (b *JapanBot) answer(args []string, s *discordgo.Session, m *discordgo.Message) {
	lastCard := b.getLatestCard(m.ChannelID)
	dictEntry := b.dictionary.IndexByID[lastCard.EntryID]

	answer := strings.Join(args[1:], " ")

	for _, sense := range dictEntry.Senses {
		for _, item := range sense.GlossaryItems {
			if strings.ToLower(item.Definition) == strings.ToLower(answer) {
				s.ChannelMessageSend(m.ChannelID, "Correct!")
				return
			}
		}
	}
	s.ChannelMessageSend(m.ChannelID, "Incorrect. Try again!")
}
