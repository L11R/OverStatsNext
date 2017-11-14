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
		profile, err := GetOverwatchProfile(info[1], info[2])
		if err != nil {
			log.Warn(err)
			text = "ERROR:\n<code>" + fmt.Sprint(err) + "</code>"
		} else {
			_, err := InsertUser(User{
				Id:      int64(update.Message.From.ID),
				Profile: profile,
				Region:  info[1],
				Nick:    info[2],
			})
			if err != nil {
				log.Warn(err)
				text = "ERROR:\n<code>" + fmt.Sprint(err) + "</code>"
			} else {
				log.Info("/save command executed successful")
				text = "Saved!"
			}
		}
	} else {
		text = "<b>Example:</b> <code>/save eu|us|kr|psn|xbl BattleTag-1337|ConsoleLogin</code> (sic, hyphen!)"
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func MeCommand(update tgbotapi.Update) {
	user, err := GetUser(update.Message.From.ID)
	if err != nil {
		log.Warn(err)
		return
	}

	place, err := GetRatingPlace(update.Message.From.ID)
	if err != nil {
		log.Warn(err)
		return
	}

	log.Info("/me command executed successful")
	text := MakeSummary(user, place)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func HeroCommand(update tgbotapi.Update) {
	user, err := GetUser(update.Message.From.ID)
	var text string

	if err != nil {
		log.Warn(err)
		text = "ERROR:\n<code>" + fmt.Sprint(err) + "</code>"
	} else {
		log.Info("/h_ command executed successful")
		hero := strings.Split(update.Message.Text, "_")[1]

		text = MakeHeroSummary(hero, user)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func RatingTopCommand(update tgbotapi.Update, platform string) {
	top, err := GetRatingTop(platform, 20)
	var text string

	if err != nil {
		log.Warn(err)
		text = "ERROR:\n<code>" + fmt.Sprint(err) + "</code>"
	} else {
		text = "<b>Rating Top:</b>\n"
		for i := range top {
			text += fmt.Sprintf("%d. %s (%d)\n", i+1, top[i].Nick, top[i].Profile.Rating)
		}
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}
