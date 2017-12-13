package main

import (
	"github.com/sirupsen/logrus"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	r "gopkg.in/gorethink/gorethink.v3"
	"os"
	"strings"
)

var log = logrus.New()

var (
	bot        *tgbotapi.BotAPI
	session    *r.Session
	dbPKPrefix = "tg:"
)

func main() {
	log.Formatter = new(logrus.TextFormatter)
	log.Info("OverStatsTelegram 1.0 started!")

	var err error

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN env variable not specified!")
	}

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	// Database pool init
	go InitConnectionPool()

	// Debug log
	bot.Debug = false

	log.Infof("authorized on account @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// userId for logger
		commandLogger := log.WithFields(logrus.Fields{"user_id": update.Message.From.ID})

		if strings.HasPrefix(update.Message.Text, "/setchat") {
			commandLogger.Info("command /setchat triggered")
			go SetChatCommand(update)
		}

		if strings.HasPrefix(update.Message.Text, "/consoletop") {
			commandLogger.Info("command /consoletop triggered")
			go RatingTopCommand(update, "console")
		}

		if strings.HasPrefix(update.Message.Text, "/pctop") {
			commandLogger.Info("command /pctop triggered")
			go RatingTopCommand(update, "pc")
		}

		// Skip all commands from groups and supergroups
		if update.Message.Chat.ID != int64(update.Message.From.ID) {
			continue
		}

		if strings.HasPrefix(update.Message.Text, "/start") {
			commandLogger.Info("command /start triggered")
			go StartCommand(update)
		}

		if strings.HasPrefix(update.Message.Text, "/donate") {
			commandLogger.Info("command /donate triggered")
			go DonateCommand(update)
		}

		if strings.HasPrefix(update.Message.Text, "/save") {
			commandLogger.Info("command /save triggered")
			go SaveCommand(update)
		}

		if strings.HasPrefix(update.Message.Text, "/me") {
			commandLogger.Info("command /me triggered")
			go MeCommand(update)
		}

		if strings.HasPrefix(update.Message.Text, "/h_") {
			commandLogger.Info("command /h_ triggered")
			go HeroCommand(update)
		}
	}
}
