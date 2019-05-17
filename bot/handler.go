package bot

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"  // for gif decoding
	_ "image/jpeg" // for jpeg decoding
	"image/png"
	"math/rand"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

var (
	// 256 colour palette
	terminalPalette color.Palette
	// regex for supported image formats
	imageRegex *regexp.Regexp
)

func init() {
	share := "./"
	if p := os.Getenv("SHARE_PATH"); p != "" {
		share = p
	}
	f, err := os.Open(path.Join(share, "colours.json"))
	if err != nil {
		panic(err)
	}
	terminalPalette, err = helpers.ReadPalette(f)
	if err != nil {
		panic(err)
	}
	imageRegex, err = regexp.Compile("\\.(png|jpeg|jpg|gif)$")
	if err != nil {
		panic(err)
	}
}

// initialises the handler map
func (b *JapanBot) createHandlerMap() HandlerMap {
	return HandlerMap{
		"analyze": b.analyse,
		"analyse": b.analyse,
		"answer":  b.answer,
		"help":    b.help,
		"hentai":  b.hentai,
		"enable":  b.enableFeature,
		"disable": b.disableFeature,
		"vape":    b.vape,
		"dither":  b.dither,
	}
}

func (b *JapanBot) analyse(args []string, s *discordgo.Session, m *discordgo.Message) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "You haven't entered a phrase!")
		return
	}

	var response string
	phrase := strings.Join(args[1:], " ")

	if helpers.IsDigits(phrase) {
		selection, err := strconv.ParseInt(phrase, 0, 0)
		if err != nil {
			panic(err)
		}
		response = b.buildSelectionResponse(int(selection), m)
	} else {
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
		b.analyseRequests[m.ChannelID] = allGrams
		response = b.buildAnalyseResponse(allGrams)
	}

	responses := strings.Split(response, "--")
	for _, r := range responses {
		_, err := s.ChannelMessageSend(
			m.ChannelID,
			r,
		)
		if err != nil {
			panic(err)
		}
	}
}

func (b *JapanBot) buildSelectionResponse(selection int, m *discordgo.Message) string {
	r, ok := b.analyseRequests[m.ChannelID]
	if !ok {
		return "You haven't specified anything to be defined!"
	}

	if selection-1 < len(r) {
		entries, ok := b.dictionary.Index[r[selection-1]]
		if ok {
			var message strings.Builder
			message.WriteString("```")
			for _, e := range entries {
				tmp := fmt.Sprintf("%s\n\n", b.buildDefinition(e, "eng"))
				lastSplit := strings.LastIndex(message.String(), "--")
				if lastSplit == -1 {
					lastSplit = 0
				}
				if (message.Len()-lastSplit)+len(tmp) >= 1996 {
					message.WriteString("\n```--```\n")
				}
				message.WriteString(tmp)
			}
			message.WriteString("```")
			return message.String()
		}
		return "No definition for this word!"
	}
	return "Definition index is invalid!"
}

func (b *JapanBot) buildDefinition(entry *jmdict.Entry, langCode string) string {
	if langCode == "" {
		langCode = "eng"
	}
	var message strings.Builder
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

func (b *JapanBot) getLastMessageWithImages(
	channelID string,
	beforeID string,
	limit int,
) (*discordgo.Message, error) {
	// get messages
	messages, err := b.session.ChannelMessages(channelID, limit, beforeID, "", "")
	if err != nil {
		return nil, err
	}
	// check each message for image attachments
	for _, m := range messages {
		for _, a := range m.Attachments {
			// if attachment is an image
			if imageRegex.MatchString(a.Filename) {
				return m, nil
			}
		}
	}
	return nil, nil
}

// image manipulation functions
type (
	imageURLFunc func(c chan<- image.Image, url string)
	imageFunc    func(c chan<- image.Image, image image.Image)
)

// manipulateImage gets the image attachments of the latest message
// and applies an imageFunc on it
// m:		message that triggered the command
// errmsg:	error message is no image is found
// status:	status message when bot is working on manipulation
// f:		the manipulation function to be applied
func (b *JapanBot) manipulateImage(
	m *discordgo.Message,
	errmsg,
	status string,
	f imageURLFunc,
) {
	// find attachments
	var attachments []*discordgo.MessageAttachment
	if len(m.Attachments) > 0 {
		for _, a := range m.Attachments {
			if imageRegex.MatchString(a.Filename) {
				attachments = m.Attachments
				break
			}
		}
	} else {
		msgs, err := b.getLastMessageWithImages(m.ChannelID, m.ID, 10)
		if err != nil {
			panic(err)
		}
		if msgs == nil {
			if errmsg != "" {
				_, err := b.session.ChannelMessageSend(m.ChannelID, errmsg)
				if err != nil {
					panic(err)
				}
			}
			return
		}
		attachments = msgs.Attachments
	}

	if status != "" {
		// post status message
		sm, err := b.session.ChannelMessageSend(m.ChannelID, status)
		if err != nil {
			panic(err)
		}
		// defer deletion of status message
		defer func() {
			if err := b.session.ChannelMessageDelete(sm.ChannelID, sm.ID); err != nil {
				panic(err)
			}
		}()
	}

	// manipulate all images
	images := make(chan image.Image)
	for _, a := range attachments {
		if imageRegex.MatchString(a.Filename) {
			go f(images, a.URL)
		}
	}

	// deliver each image
	for img := range images {
		var bs bytes.Buffer
		if err := png.Encode(&bs, img); err != nil {
			panic(err)
		}
		_, err := b.session.ChannelFileSend(
			m.ChannelID,
			fmt.Sprintf("%x%x.png", rand.Int(), rand.Int()),
			&bs,
		)
		if err != nil {
			panic(err)
		}
	}
}

func (b *JapanBot) vape(args []string, s *discordgo.Session, m *discordgo.Message) {
	b.manipulateImage(
		m,
		"Couldn't find an image to vape!",
		"Working on it!",
		vapeImageURL,
	)
}

func (b *JapanBot) dither(args []string, s *discordgo.Session, m *discordgo.Message) {
	b.manipulateImage(
		m,
		"Couldn't find an image to dither!",
		"Dithering now!",
		ditherImageURL,
	)
}

func vapeImageURL(c chan<- image.Image, url string) {
	r, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	if img, _, err := image.Decode(r.Body); err != nil {
		panic(err)
	} else {
		go vapeImage(c, img)
	}
}

func vapeImage(c chan<- image.Image, img image.Image) {
	bayerMatrix := helpers.BayerMatrix16
	l := len(bayerMatrix)
	newImg := image.NewRGBA(img.Bounds())
	size := img.Bounds().Size()
	var wg sync.WaitGroup
	for y := 0; y < size.Y; y++ {
		// do each row concurrently
		wg.Add(1)
		go func(y int) {
			// get and sort row
			row := make([]color.Color, size.X)
			for x := 0; x < size.X; x++ {
				row[x] = helpers.Neonify(
					img.At(x, y),
					float32(x)/float32(size.X)*0.7,
				)
			}
			row = helpers.SortColours(row)
			// place row in newImg
			for x, c := range row {
				newImg.Set(
					x,
					y,
					helpers.Dither(
						c,
						terminalPalette,
						bayerMatrix[x%l][y%l],
					),
				)
			}
			wg.Done()
		}(y)
	}
	wg.Wait()
	c <- newImg
}

func ditherImageURL(c chan<- image.Image, url string) {
	r, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	if img, _, err := image.Decode(r.Body); err != nil {
		panic(err)
	} else {
		go ditherImage(c, img)
	}
}

func ditherImage(c chan<- image.Image, img image.Image) {
	bayerMatrix := helpers.BayerMatrix16
	l := len(bayerMatrix)
	newImg := image.NewRGBA(img.Bounds())
	size := img.Bounds().Size()
	var wg sync.WaitGroup
	for y := 0; y < size.Y; y++ {
		// do each row concurrently
		wg.Add(1)
		go func(y int) {
			for x := 0; x < size.X; x++ {
				newImg.Set(
					x,
					y,
					helpers.Dither(
						img.At(x, y),
						terminalPalette,
						bayerMatrix[x%l][y%l],
					),
				)
			}
			wg.Done()
		}(y)
	}
	wg.Wait()
	c <- newImg
}
