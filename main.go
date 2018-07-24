package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hakasec/japanbot-go/bot"
	"github.com/hakasec/japanbot-go/bot/config"
)

var botConfig *config.BotConfiguration

func init() {
	var err error
	botConfig, err = config.LoadFromFile("./config.json")
	if err != nil {
		panic(err)
	}
}

func main() {
	bot, err := bot.New(botConfig)
	if err != nil {
		panic(err)
	}

	err = bot.Start()
	if err != nil {
		panic(err)
	}
	defer bot.Stop()

	fmt.Println("Bot is running!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
