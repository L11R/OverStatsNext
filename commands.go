package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

func StartCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Use /help")
	msg.ParseMode = "HTML"
	bot.Send(msg)

	log.Info("/start command executed successful")
}

func HelpCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "<code>/save eu|us|kr BattleTag-1337</code> — save your profile into DB\n"+
		"<code>/me</code> — getting saved profile\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)

	log.Info("/help command executed successful")
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

func GetCommand(update tgbotapi.Update) {
	info := strings.Split(update.Message.Text, " ")
	var text string

	if len(info) == 3 {
		profile, err := GetOverwatchProfile(info[1], info[2])
		if err != nil {
			log.Warn(err)
			text = "ERROR:\n<code>" + fmt.Sprint(err) + "</code>"
		} else {
			log.Info("/get command executed successful")
			text = MakeSummary(profile)
		}
	} else {
		text = "Not enough arguments!"
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
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
				int64(update.Message.From.ID),
				profile,
				info[2],
				info[1],
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
		text = "Not enough arguments!"
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func MeCommand(update tgbotapi.Update) {
	user, err := GetUser(update.Message.From.ID)
	var text string

	if err != nil {
		log.Warn(err)
		text = "ERROR:\n<code>" + fmt.Sprint(err) + "</code>"
	} else {
		log.Info("/me command executed successful")
		text = MakeSummary(user.Profile)
	}

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

		text = MakeHeroSummary(hero, user.Profile)
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

func SessionReport(change Change) {
	// Check OldVal and NewOld existing
	if change.OldVal.Profile != nil && change.NewVal.Profile != nil {
		oldStats := Report{
			Rating: change.OldVal.Profile.Rating,
			Level:  change.OldVal.Profile.Prestige*100 + change.OldVal.Profile.Level,
		}

		if competitiveStats, ok := change.OldVal.Profile.CompetitiveStats.CareerStats["allHeroes"]; ok {
			if gamesPlayed, ok := competitiveStats.Game["gamesPlayed"]; ok {
				oldStats.Games = int(gamesPlayed.(float64))
			}
			if gamesWon, ok := competitiveStats.Game["gamesWon"]; ok {
				oldStats.Wins = int(gamesWon.(float64))
			}
			if gamesTied, ok := competitiveStats.Miscellaneous["gamesTied"]; ok {
				oldStats.Ties = int(gamesTied.(float64))
			}
			if gamesLost, ok := competitiveStats.Miscellaneous["gamesLost"]; ok {
				oldStats.Losses = int(gamesLost.(float64))
			}
		}

		newStats := Report{
			Rating: change.NewVal.Profile.Rating,
			Level:  change.NewVal.Profile.Prestige*100 + change.NewVal.Profile.Level,
		}

		if competitiveStats, ok := change.NewVal.Profile.CompetitiveStats.CareerStats["allHeroes"]; ok {
			if gamesPlayed, ok := competitiveStats.Game["gamesPlayed"]; ok {
				newStats.Games = int(gamesPlayed.(float64))
			}
			if gamesWon, ok := competitiveStats.Game["gamesWon"]; ok {
				newStats.Wins = int(gamesWon.(float64))
			}
			if gamesTied, ok := competitiveStats.Miscellaneous["gamesTied"]; ok {
				newStats.Ties = int(gamesTied.(float64))
			}
			if gamesLost, ok := competitiveStats.Miscellaneous["gamesLost"]; ok {
				newStats.Losses = int(gamesLost.(float64))
			}
		}

		diffStats := Report{
			newStats.Rating - oldStats.Rating,
			newStats.Level - oldStats.Level,
			newStats.Games - oldStats.Games,
			newStats.Wins - oldStats.Wins,
			newStats.Ties - oldStats.Ties,
			newStats.Losses - oldStats.Losses,
		}

		if diffStats.Games != 0 {
			log.Infof("sending report to %d", change.NewVal.Id)
			text := "<b>Session Report</b>\n\n"

			text += AddInfo("Rating", oldStats.Rating, newStats.Rating, diffStats.Rating)
			text += AddInfo("Wins", oldStats.Wins, newStats.Wins, diffStats.Wins)
			text += AddInfo("Losses", oldStats.Losses, newStats.Losses, diffStats.Losses)
			text += AddInfo("Ties", oldStats.Ties, newStats.Ties, diffStats.Ties)
			text += AddInfo("Level", oldStats.Level, newStats.Level, diffStats.Level)

			msg := tgbotapi.NewMessage(change.NewVal.Id, text)
			msg.ParseMode = "HTML"
			bot.Send(msg)
		}
	}
}
