package main

import (
	"errors"
	"fmt"
	"github.com/jasonlvhit/gocron"
	"github.com/sdwolfe32/ovrstat/ovrstat"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func UpdateProfile(id int64, region string, nick string) {
	profile, err := GetOverwatchProfile(region, nick)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id": id,
		}).Warn(err)
	} else {
		_, err := UpdateUser(User{
			Id:      id,
			Profile: profile,
		})
		if err != nil {
			log.WithFields(logrus.Fields{
				"id": id,
			}).Warn(err)
		} else {
			log.WithFields(logrus.Fields{
				"id": id,
			}).Infof("%s (%s) profile updated!", nick, region)
		}
	}

	defer wg.Done()
}

func SplitUsers(users []UserWithoutProfile, lim int) [][]UserWithoutProfile {
	var chunk []UserWithoutProfile
	chunks := make([][]UserWithoutProfile, 0, len(users)/lim+1)

	for len(users) >= lim {
		chunk, users = users[:lim], users[lim:]
		chunks = append(chunks, chunk)
	}

	if len(users) > 0 {
		chunks = append(chunks, users[:])
	}

	return chunks
}

func UpdateAllProfiles() {
	users, err := GetUsersWithoutProfile()
	if err != nil {
		log.Warn(err)
	} else {
		users := SplitUsers(users, 50)

		for _, group := range users {
			wg.Add(len(group))

			for _, user := range group {
				go UpdateProfile(user.Id, user.Region, user.Nick)
			}

			wg.Wait()
		}
	}
}

func InitCron() {
	gocron.Every(1).Minutes().Do(UpdateAllProfiles)
	//gocron.Every(10).Seconds().Do(UpdateAllProfiles)
	<-gocron.Start()
}

// Make small text summary based on profile
func MakeSummary(user User) string {
	profile := user.Profile
	text := fmt.Sprintf("<b>%s</b> (<b>%d</b> sr / <b>%d</b> lvl)\n", profile.Name, profile.Rating, profile.Prestige*100+profile.Level)

	if careerStats, ok := profile.CompetitiveStats.CareerStats["allHeroes"]; ok {
		var stats Report
		if gamesPlayed, ok := careerStats.Game["gamesPlayed"]; ok {
			stats.Games = int(gamesPlayed.(float64))
		}
		if gamesWon, ok := careerStats.Game["gamesWon"]; ok {
			stats.Wins = int(gamesWon.(float64))
		}
		if gamesTied, ok := careerStats.Miscellaneous["gamesTied"]; ok {
			stats.Ties = int(gamesTied.(float64))
		}
		if gamesLost, ok := careerStats.Miscellaneous["gamesLost"]; ok {
			stats.Losses = int(gamesLost.(float64))
		}

		text += fmt.Sprintf("%d-%d-%d / <b>%0.2f%%</b> winrate\n", stats.Wins, stats.Losses, stats.Ties, float64(stats.Wins)/float64(stats.Games)*100)

		// Temp struct for k/d counting
		type KD struct {
			Eliminations float64
			Deaths       float64
			Ratio        float64
		}

		var kd KD

		if eliminations, ok := careerStats.Combat["eliminations"]; ok {
			kd.Eliminations = eliminations.(float64)
		}
		if deaths, ok := careerStats.Deaths["deaths"]; ok {
			kd.Deaths = deaths.(float64)
		}

		if kd.Deaths > 0 {
			kd.Ratio = kd.Eliminations / kd.Deaths
			text += fmt.Sprintf("<b>%0.2f</b> k/d\n\n", kd.Ratio)
		}

		text += "<b>7 top played heroes:</b>\n"
		var topPlayedHeroes Heroes
		for name, elem := range profile.CompetitiveStats.TopHeroes {
			topPlayedHeroes = append(topPlayedHeroes, Hero{
				Name:                name,
				TimePlayedInSeconds: elem.TimePlayedInSeconds,
			})
		}

		// Sort top played heroes in descending
		sort.Sort(sort.Reverse(topPlayedHeroes))

		for i := 0; i < 7; i++ {
			text += fmt.Sprintf(
				"%s (%s) /h_%s\n",
				strings.Title(strings.ToLower(topPlayedHeroes[i].Name)),
				profile.CompetitiveStats.TopHeroes[topPlayedHeroes[i].Name].TimePlayed,
				topPlayedHeroes[i].Name,
			)
		}
	}

	text += fmt.Sprint("\n<b>Last Updated:</b>\n", user.Date.Format("15:04:05 / 02.01.2006 MST"))

	return text
}

func MakeHeroSummary(hero string, user User) string {
	profile := user.Profile
	text := fmt.Sprintf("<b>%s</b>", strings.Title(strings.ToLower(hero)))

	if heroStats, ok := profile.CompetitiveStats.CareerStats[hero]; ok {
		if heroAdditionalStats, ok := profile.CompetitiveStats.TopHeroes[hero]; ok {
			text += fmt.Sprintf(" (%s)\n", heroAdditionalStats.TimePlayed)

			// Temp struct for winrate counting
			type WR struct {
				Wins  float64
				Games float64
				Ratio float64
			}

			var wr WR

			if gamesWon, ok := heroStats.Game["gamesWon"]; ok {
				wr.Wins = gamesWon.(float64)
			}
			if gamesPlayed, ok := heroStats.Game["gamesPlayed"]; ok {
				wr.Games = gamesPlayed.(float64)
			}

			if wr.Games > 0 {
				wr.Ratio = wr.Wins / wr.Games * 100
				text += fmt.Sprintf("<b>%0.2f%%</b> hero winrate\n", wr.Ratio)
			}

			if eliminations, ok := heroStats.Combat["eliminations"]; ok {
				eliminationsPerMin := eliminations.(float64) / (float64(heroAdditionalStats.TimePlayedInSeconds) / 60)
				text += fmt.Sprintf("<b>%0.2f</b> eliminations per min\n", eliminationsPerMin)
			}

			if eliminationsPerLife, ok := heroStats.Combat["eliminationsPerLife"]; ok {
				text += fmt.Sprintf("<b>%0.2f</b> k/d ratio\n", eliminationsPerLife)
			}

			if accuracy, ok := heroStats.Combat["weaponAccuracy"]; ok {
				text += fmt.Sprintf("<b>%s</b> accuracy\n", accuracy)
			}

			if damageDone, ok := heroStats.Combat["damageDone"]; ok {
				damagePerMin := damageDone.(float64) / (float64(heroAdditionalStats.TimePlayedInSeconds) / 60)
				text += fmt.Sprintf("<b>%0.0f</b> damage per min\n", damagePerMin)
			}

			if blocked, ok := heroStats.Miscellaneous["damageBlocked"]; ok {
				blockedPerMin := blocked.(float64) / (float64(heroAdditionalStats.TimePlayedInSeconds) / 60)
				text += fmt.Sprintf("<b>%0.0f</b> blocked per min\n", blockedPerMin)
			}

			if healing, ok := heroStats.Miscellaneous["healingDone"]; ok {
				healingPerMin := healing.(float64) / (float64(heroAdditionalStats.TimePlayedInSeconds) / 60)
				text += fmt.Sprintf("<b>%0.0f</b> healing per min\n", healingPerMin)
			}

			if objKills, ok := heroStats.Combat["objectiveKills"]; ok {
				objKillsPerMin := objKills.(float64) / (float64(heroAdditionalStats.TimePlayedInSeconds) / 60)
				text += fmt.Sprintf("<b>%0.2f</b> obj. kills per min\n", objKillsPerMin)
			}

			if crits, ok := heroStats.Combat["criticalHits"]; ok {
				critsPerMin := crits.(float64) / (float64(heroAdditionalStats.TimePlayedInSeconds) / 60)
				text += fmt.Sprintf("<b>%0.2f</b> crits per min\n", critsPerMin)
			}
		} else {
			text += "\nNOT AVAILABLE"
		}
	} else {
		text += "\nNOT AVAILABLE"
	}

	text += fmt.Sprint("\n<b>Last Updated:</b>\n", user.Date.Format("15:04:05 / 02.01.2006 MST"))

	return text
}

//
func AddInfo(name string, oldInfo int, newInfo int, diffInfo int) string {
	text := fmt.Sprintf("%s:\n<code>%d | %d |", name, oldInfo, newInfo)
	if diffInfo > 0 {
		text += fmt.Sprintf(" +%d 📈\n</code>", diffInfo)
	} else if diffInfo == 0 {
		text += fmt.Sprintf(" %d —\n</code>", diffInfo)
	} else {
		text += fmt.Sprintf(" %d 📉\n</code>", diffInfo)
	}

	return text
}

// Fetch Overwatch profile based on region and BattleTag / PSN ID / Xbox Live Account
func GetOverwatchProfile(region string, nick string) (*ovrstat.PlayerStats, error) {
	if region == "eu" || region == "us" || region == "kr" {
		return ovrstat.PCStats(region, nick)
	} else if region == "psn" || region == "xbl" {
		return ovrstat.ConsoleStats(region, nick)
	}

	return nil, errors.New("region is wrong")
}
