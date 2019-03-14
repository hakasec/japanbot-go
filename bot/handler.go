package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hakasec/japanbot-go/bot/database/models"
	"github.com/hakasec/japanbot-go/bot/helpers"
	jmdict "github.com/hakasec/jmdict-go"
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

	if helpers.IsDigits(phrase) {
		selection, err := strconv.ParseInt(phrase, 0, 0)
		if err != nil {
			panic(err)
		}
		go b.handleAnalyseSelection(int(selection), s, m)
		return
	}

	var allGrams []string
	// generate a list of ngrams of size 1 through len(phrase)
	for ngramSize := 1; ngramSize <= len(phrase); ngramSize++ {
		tmpNgram := helpers.CreateNgrams(phrase, ngramSize)
		for _, gram := range tmpNgram {
			// remove spaces
			gram = strings.Replace(gram, " ", "", -1)
			if gram == "" {
				continue
			}
			// check if already in list
			if !helpers.StringSliceContains(allGrams, gram) {
				// check for definition
				if _, ok := b.dictionary.Index[gram]; ok {
					// add to list
					allGrams = append(allGrams, gram)
				}
			}
		}
	}

	responses := strings.Split(b.buildAnalyseResponse(allGrams), "--")
	for _, r := range responses {
		_, err := s.ChannelMessageSend(
			m.ChannelID,
			r,
		)
		if err != nil {
			panic(err)
		}
	}

	b.analyseRequests[m.ChannelID] = allGrams
}

func (b *JapanBot) handleAnalyseSelection(selection int, s *discordgo.Session, m *discordgo.Message) {
	r, ok := b.analyseRequests[m.ChannelID]
	if !ok {
		s.ChannelMessageSend(m.ChannelID, "You haven't specified anything to be defined!")
	} else {
		if selection-1 < len(r) {
			entries, ok := b.dictionary.Index[r[selection-1]]
			if ok {
				for _, e := range entries {
					s.ChannelMessageSend(m.ChannelID, b.buildDefinition(e, "eng"))
				}
			} else {
				s.ChannelMessageSend(m.ChannelID, "No definition for this word!")
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "Definition index is invalid!")
		}
	}
}

func (b *JapanBot) buildDefinition(entry *jmdict.Entry, langCode string) string {
	if langCode == "" {
		langCode = "eng"
	}
	var message strings.Builder
	message.WriteString("```")
	for _, reading := range entry.KanjiElements {
		message.WriteString(fmt.Sprintln(reading.Phrase))
	}
	for _, reading := range entry.ReadingElements {
		if reading.PhraseNoKanji != "" {
			message.WriteString(
				fmt.Sprintf("%s (%s)\n", reading.Phrase, reading.PhraseNoKanji),
			)
		} else {
			message.WriteString(fmt.Sprintln(reading.Phrase))
		}
	}
	message.WriteString("\n")
	for _, sense := range entry.Senses {
		for _, gloss := range sense.GlossaryItems {
			language := gloss.Language
			if language == "" {
				language = "eng"
			}
			if language == langCode {
				message.WriteString(
					fmt.Sprintf("%s\n", gloss.Definition),
				)
			}
		}
	}
	message.WriteString("```")
	return message.String()
}

func (b *JapanBot) buildAnalyseResponse(ngrams []string) string {
	if len(ngrams) == 0 {
		return "No definitions found :("
	}

	var message strings.Builder
	message.WriteString("```\nPick a phrase:\n")
	width := helpers.GetNumDigits(len(ngrams))
	for i, gram := range ngrams {
		tmp := fmt.Sprintf(
			"%d: %s%s\n",
			i+1,
			strings.Repeat(" ", width-helpers.GetNumDigits(i+1)),
			gram,
		)
		lastSplit := strings.LastIndex(message.String(), "--")
		if lastSplit == -1 {
			lastSplit = 0
		}
		// check for length, if over 1993 in len, add slip character
		if (message.Len()-lastSplit)+len(tmp) >= 1996 {
			message.WriteString("\n```--```\n")
		}
		message.WriteString(tmp)
	}
	message.WriteString(
		fmt.Sprintf("\nUse jpn!analyse [1-%d]\n", len(ngrams)),
	)
	message.WriteString("```")
	return message.String()
}

func (b *JapanBot) help(args []string, s *discordgo.Session, m *discordgo.Message) {
	var builder strings.Builder
	builder.WriteString("```\n")

	builder.WriteString(
		`All commands begin with "jpn!"

Available Commands:

- analyse/analyze: Analyse a Japanese sentence. 

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
	dictEntries := b.dictionary.IndexByID[lastCard.EntryID]

	answer := strings.Join(args[1:], " ")

	for _, dictEntry := range dictEntries {
		for _, sense := range dictEntry.Senses {
			for _, item := range sense.GlossaryItems {
				if strings.ToLower(item.Definition) == strings.ToLower(answer) {
					s.ChannelMessageSend(m.ChannelID, "Correct!")
					return
				}
			}
		}
	}
	s.ChannelMessageSend(m.ChannelID, "Incorrect. Try again!")
}
