package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

func StartCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Simple bot for Overwatch by @kraso\n\n"+
		"<b>How to use:</b>\n"+
		"1. Use /save to save your game profile.\n"+
		"2. Use /me to see your stats.\n"+
		"3. ???\n"+
		"4. PROFIT!\n\n"+
		"<b>Features:</b>\n"+
		"— Player profile (/me command)\n"+
		"— Small summary for heroes\n"+
		"— Reports after every game session\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)

	log.Info("/start command executed successful")
}

func DonateCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "If you find this bot helpful, "+
		"<a href=\"https://paypal.me/krasovsky\">you can make small donation</a> to help me pay server bills!")
	msg.ParseMode = "HTML"
	bot.Send(msg)

	log.Info("donate command executed successful")
}

type Hero struct {
	Name                string
	TimePlayedInSeconds int
}

type Heroes []Hero

func (hero Heroes) Len() int {
	return len(hero)
}

func (hero Heroes) Less(i, j int) bool {
	return hero[i].TimePlayedInSeconds < hero[j].TimePlayedInSeconds
}

func (hero Heroes) Swap(i, j int) {
	hero[i], hero[j] = hero[j], hero[i]
}

func SaveCommand(update tgbotapi.Update) {
	info := strings.Split(update.Message.Text, " ")
	var text string

	if len(info) == 3 {
		if info[1] != "psn" && info[1] != "xbl" {
			info[2] = strings.Replace(info[2], "#", "-", -1)
		}

		profile, err := GetOverwatchProfile(info[1], info[2])
		if err != nil {
			log.Warn(err)
			text = "Player not found!"
		} else {
			_, err := InsertUser(User{
				Id:      fmt.Sprint(dbPKPrefix, update.Message.From.ID),
				Profile: profile,
				Region:  info[1],
				Nick:    info[2],
			})
			if err != nil {
				log.Warn(err)
				return
			}

			log.Info("/save command executed successful")
			text = "Saved!"
		}
	} else {
		text = "<b>Example:</b> <code>/save eu|us|kr|psn|xbl BattleTag#1337|ConsoleLogin</code>"
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func MeCommand(update tgbotapi.Update) {
	user, err := GetUser(fmt.Sprint(dbPKPrefix, update.Message.From.ID))
	if err != nil {
		log.Warn(err)
		return
	}

	place, err := GetRatingPlace(fmt.Sprint(dbPKPrefix, update.Message.From.ID))
	if err != nil {
		log.Warn(err)
		return
	}

	log.Info("/me command executed successful")

	var text string
	info := strings.Split(update.Message.Text, "_")

	if len(info) == 1 {
		text = MakeSummary(user, place, "CompetitiveStats")
	} else if len(info) == 2 && info[1] == "quick" {
		text = MakeSummary(user, place, "QuickPlayStats")
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func HeroCommand(update tgbotapi.Update) {
	user, err := GetUser(fmt.Sprint(dbPKPrefix, update.Message.From.ID))
	if err != nil {
		log.Warn(err)
		return
	}

	log.Info("/h_ command executed successful")

	var text string
	info := strings.Split(update.Message.Text, "_")
	hero := info[1]

	if len(info) == 2 {
		text = MakeHeroSummary(hero, "CompetitiveStats", user)
	} else if len(info) == 3 && info[2] == "quick" {
		text = MakeHeroSummary(hero, "QuickPlayStats", user)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func RatingTopCommand(update tgbotapi.Update, platform string) {
	var chatId int64
	if update.Message.Chat.Type == "private" {
		chatId = 0
	} else {
		chatId = update.Message.Chat.ID
	}

	top, err := GetRatingTop(platform, 20, chatId)
	if err != nil {
		log.Warn(err)
		return
	}

	text := "<b>Rating Top:</b>\n"
	for i, elem := range top {
		nick := elem.Patreon + elem.Nick
		if elem.Region != "psn" && elem.Region != "xbl" {
			nick = strings.Replace(nick, "-", "#", -1)
		}
		text += fmt.Sprintf("%d. %s (%d)\n", i+1, nick, elem.Profile.Rating)
	}
	if len(top) == 0 {
		text += "It's empty..."
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func SetChatCommand(update tgbotapi.Update) {
	// Skip if it's private
	if update.Message.Chat.Type == "private" {
		return
	}

	res, err := UpdateUser(User{
		Id:   fmt.Sprint("tg:", update.Message.From.ID),
		Chat: update.Message.Chat.ID,
	})
	if err != nil {
		log.Warn(err)
		return
	}

	var text string
	if res.Unchanged != 0 {
		text = "<b>Error:</b> This chat already set as primary!"
	}
	if res.Replaced != 0 || res.Updated != 0 {
		text = "<b>Done:</b> Set as primary chat!"
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}
