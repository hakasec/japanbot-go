package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/hakasec/japanbot-go/bot"
	"github.com/hakasec/japanbot-go/bot/config"
)

var botConfig *config.BotConfiguration

func init() {
	cpath := "./"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		cpath = p
	}
	var err error
	botConfig, err = config.LoadFromFile(
		path.Join(cpath, "config.json"),
	)
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
