package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
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
		"analyse": b.analyse,
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
			message := fmt.Sprintf("```\n%s:\n", gram)
			hasDefinition := false
			for _, sense := range entry.Senses {
				for _, gloss := range sense.GlossaryItems {
					language := gloss.Language
					if language == "" {
						language = "eng"
					}
					if language == commandLang {
						hasDefinition = true
						message += fmt.Sprintf("\t%s\n", gloss.Definition)
					}
				}
			}
			message += "```"
			if hasDefinition {
				hasContent = true
				s.ChannelMessageSend(m.ChannelID, message)
			}
		}
	}
	if !hasContent {
		s.ChannelMessageSend(m.ChannelID, "ごめん! No definitions found!")
	}
}
