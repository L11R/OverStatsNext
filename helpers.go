package main

import (
	"errors"
	"fmt"
	"github.com/sdwolfe32/ovrstat/ovrstat"
	r "gopkg.in/gorethink/gorethink.v3"
	"sort"
	"strings"
)

// Make small text summary based on profile
func MakeSummary(user User, top Top, mode string) string {
	text := fmt.Sprintf("<b>%s</b> (<b>%d</b> sr / <b>%d</b> lvl)\n", user.Profile.Name, user.Profile.Rating, user.Profile.Prestige*100+user.Profile.Level)

	var stats ovrstat.StatsCollection
	if mode == "CompetitiveStats" {
		stats = user.Profile.CompetitiveStats
	}
	if mode == "QuickPlayStats" {
		stats = user.Profile.QuickPlayStats
	}

	if careerStats, ok := stats.CareerStats["allHeroes"]; ok {
		var basicStats Report
		if gamesPlayed, ok := careerStats.Game["gamesPlayed"]; ok {
			basicStats.Games = int(gamesPlayed.(float64))
		}
		if gamesWon, ok := careerStats.Game["gamesWon"]; ok {
			basicStats.Wins = int(gamesWon.(float64))
		}
		if gamesTied, ok := careerStats.Miscellaneous["gamesTied"]; ok {
			basicStats.Ties = int(gamesTied.(float64))
		}
		if gamesLost, ok := careerStats.Miscellaneous["gamesLost"]; ok {
			basicStats.Losses = int(gamesLost.(float64))
		}

		if mode == "CompetitiveStats" {
			text += fmt.Sprintf("%d-%d-%d / <b>%0.2f%%</b> winrate\n", basicStats.Wins, basicStats.Losses, basicStats.Ties, float64(basicStats.Wins)/float64(basicStats.Games)*100)
		} else if mode == "QuickPlayStats" {
			text += fmt.Sprintf("<b>%d</b> wins\n", basicStats.Wins)
		}

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

		if mode == "CompetitiveStats" {
			text += fmt.Sprintf("<b>Rating Top:</b>\n#%d (%0.2f%%)\n\n", top.Place, top.Rank)
		}

		text += "<b>7 top played heroes:</b>\n"
		var topPlayedHeroes Heroes
		for name, elem := range stats.TopHeroes {
			topPlayedHeroes = append(topPlayedHeroes, Hero{
				Name:                name,
				TimePlayedInSeconds: elem.TimePlayedInSeconds,
			})
		}

		// Sort top played heroes in descending
		sort.Sort(sort.Reverse(topPlayedHeroes))

		for i := 0; i < 7; i++ {
			var format string
			if mode == "CompetitiveStats" {
				format = fmt.Sprint("%s (%s) /h_%s\n")
			} else if mode == "QuickPlayStats" {
				format = fmt.Sprint("%s (%s) /h_%s_quick\n")
			}

			text += fmt.Sprintf(
				format,
				strings.Title(strings.ToLower(topPlayedHeroes[i].Name)),
				stats.TopHeroes[topPlayedHeroes[i].Name].TimePlayed,
				topPlayedHeroes[i].Name,
			)
		}
	}

	text += fmt.Sprint("\n<b>Last Updated:</b>\n", user.Date.Format("15:04:05 / 02.01.2006 MST"))

	return text
}

func MakeHeroSummary(hero string, mode string, user User) string {
	text := fmt.Sprintf("<b>%s</b>", strings.Title(strings.ToLower(hero)))

	var stats ovrstat.StatsCollection
	if mode == "CompetitiveStats" {
		stats = user.Profile.CompetitiveStats
	}
	if mode == "QuickPlayStats" {
		stats = user.Profile.QuickPlayStats
	}

	if heroStats, ok := stats.CareerStats[hero]; ok {
		if heroAdditionalStats, ok := stats.TopHeroes[hero]; ok {
			text += fmt.Sprintf(" (%s)\n", heroAdditionalStats.TimePlayed)
			if cards, ok := heroStats.MatchAwards["cards"]; ok {
				text += fmt.Sprintf("🃏%0.0f ", cards)
			}
			if medalsGold, ok := heroStats.MatchAwards["medalsGold"]; ok {
				text += fmt.Sprintf("🥇%0.0f ", medalsGold)
			}
			if medalsSilver, ok := heroStats.MatchAwards["medalsSilver"]; ok {
				text += fmt.Sprintf("🥈%0.0f ", medalsSilver)
			}
			if medalsBronze, ok := heroStats.MatchAwards["medalsBronze"]; ok {
				text += fmt.Sprintf("🥉%0.0f ", medalsBronze)
			}

			text += "\n"

			if mode == "CompetitiveStats" {
				text += fmt.Sprintf("<b>%d%%</b> hero winrate", heroAdditionalStats.WinPercentage)

				res, err := GetRank(
					fmt.Sprint(dbPKPrefix, user.Id),
					r.Row.Field("profile").Field(mode).Field("TopHeroes").Field(hero).Field("WinPercentage"),
				)
				if err != nil {
					text += fmt.Sprint(" (error)\n")
				} else {
					text += fmt.Sprintf(" (#%d, %0.0f%%)\n", res.Place, res.Rank)
				}
			}

			if eliminationsPerLife, ok := heroStats.Combat["eliminationsPerLife"]; ok {
				text += fmt.Sprintf("<b>%0.2f</b> k/d ratio", eliminationsPerLife)

				res, err := GetRank(
					fmt.Sprint(dbPKPrefix, user.Id),
					r.Row.Field("profile").Field(mode).Field("CareerStats").Field(hero).Field("Combat").Field("eliminationsPerLife"),
				)
				if err != nil {
					text += fmt.Sprint(" (error)\n")
				} else {
					text += fmt.Sprintf(" (#%d, %0.0f%%)\n", res.Place, res.Rank)
				}
			}

			if accuracy, ok := heroStats.Combat["weaponAccuracy"]; ok {
				text += fmt.Sprintf("<b>%s</b> accuracy", accuracy)

				res, err := GetRank(
					fmt.Sprint(dbPKPrefix, user.Id),
					r.Row.Field("profile").Field(mode).Field("CareerStats").Field(hero).Field("Combat").Field("weaponAccuracy"),
				)
				if err != nil {
					text += fmt.Sprint(" (error)\n")
				} else {
					text += fmt.Sprintf(" (#%d, %0.0f%%)\n", res.Place, res.Rank)
				}
			}

			timePlayedInMinutes := float64(heroAdditionalStats.TimePlayedInSeconds) / 60

			if eliminations, ok := heroStats.Combat["eliminations"]; ok {
				eliminationsPerMin := eliminations.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("<b>%0.2f</b> eliminations per min\n", eliminationsPerMin)
			}

			if damageDone, ok := heroStats.Combat["damageDone"]; ok {
				damagePerMin := damageDone.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("<b>%0.0f</b> damage per min\n", damagePerMin)
			}

			if blocked, ok := heroStats.Miscellaneous["damageBlocked"]; ok {
				blockedPerMin := blocked.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("<b>%0.0f</b> blocked per min\n", blockedPerMin)
			}

			if healing, ok := heroStats.Miscellaneous["healingDone"]; ok {
				healingPerMin := healing.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("<b>%0.0f</b> healing per min\n", healingPerMin)
			}

			if objKills, ok := heroStats.Combat["objectiveKills"]; ok {
				objKillsPerMin := objKills.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("<b>%0.2f</b> obj. kills per min\n", objKillsPerMin)
			}

			if crits, ok := heroStats.Combat["criticalHits"]; ok {
				critsPerMin := crits.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("<b>%0.2f</b> crits per min\n", critsPerMin)
			}

			// HERO SPECIFIC
			text += "\n<b>Hero Specific:</b>\n"
			switch hero {
			case "ana":
				if scopedAccuracy, ok := heroStats.HeroSpecific["scopedAccuracy"]; ok {
					text += fmt.Sprintf("<b>%s</b> scoped accuracy\n", scopedAccuracy)
				}
				if enemiesSlept, ok := heroStats.Miscellaneous["enemiesSlept"]; ok {
					enemiesSleptPerMin := enemiesSlept.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> enemies slept per min\n", enemiesSleptPerMin)
				}
			case "bastion":
				if reconKills, ok := heroStats.HeroSpecific["reconKills"]; ok {
					reconKillsPerMin := reconKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> recon kills per min\n", reconKillsPerMin)
				}
				if sentryKills, ok := heroStats.HeroSpecific["sentryKills"]; ok {
					sentryKillsPerMin := sentryKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> sentry kills per min\n", sentryKillsPerMin)
				}
				if tankKills, ok := heroStats.HeroSpecific["tankKills"]; ok {
					tankKillsPerMin := tankKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> tank kills per min\n", tankKillsPerMin)
				}
			case "dVa":
				if blocked, ok := heroStats.HeroSpecific["damageBlocked"]; ok {
					blockedPerMin := blocked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> blocked per min\n", blockedPerMin)
				}
				if mechsCalled, ok := heroStats.HeroSpecific["mechsCalled"]; ok {
					mechsCalledPerMin := mechsCalled.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> mechs called per min\n", mechsCalledPerMin)
				}
				if mechDeaths, ok := heroStats.HeroSpecific["mechDeaths"]; ok {
					mechDeathsPerMin := mechDeaths.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> mech deaths per min\n", mechDeathsPerMin)
				}
				if selfDestructKills, ok := heroStats.Miscellaneous["selfDestructKills"]; ok {
					selfDestructKillsPerMin := selfDestructKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> self destruct kills per min\n", selfDestructKillsPerMin)
				}
			case "doomfist":
				if abilityDamageDone, ok := heroStats.HeroSpecific["abilityDamageDone"]; ok {
					abilityDamageDonePerMin := abilityDamageDone.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> ability damage done per min\n", abilityDamageDonePerMin)
				}
				if meteorStrikeKills, ok := heroStats.HeroSpecific["meteorStrikeKills"]; ok {
					meteorStrikeKillsPerMin := meteorStrikeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> meteor strike kills per min\n", meteorStrikeKillsPerMin)
				}
				if shieldsCreated, ok := heroStats.HeroSpecific["shieldsCreated"]; ok {
					shieldsCreatedPerMin := shieldsCreated.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> shields created per min\n", shieldsCreatedPerMin)
				}
			case "genji":
				if damageReflected, ok := heroStats.HeroSpecific["damageReflected"]; ok {
					damageReflectedPerMin := damageReflected.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> damage reflected per min\n", damageReflectedPerMin)
				}
				if dragonbladesKills, ok := heroStats.HeroSpecific["dragonbladesKills"]; ok {
					dragonbladesKillsPerMin := dragonbladesKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> dragonblades kills per min\n", dragonbladesKillsPerMin)
				}
			case "hanzo":
				if dragonstrikeKills, ok := heroStats.HeroSpecific["dragonstrikeKills"]; ok {
					dragonstrikeKillsPerMin := dragonstrikeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> dragonstrike kills per min\n", dragonstrikeKillsPerMin)
				}
				if scatterArrowKills, ok := heroStats.HeroSpecific["scatterArrowKills"]; ok {
					scatterArrowKillsPerMin := scatterArrowKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> scatter arrow kills per min\n", scatterArrowKillsPerMin)
				}
			case "junkrat":
				if enemiesTrapped, ok := heroStats.HeroSpecific["enemiesTrapped"]; ok {
					enemiesTrappedPerMin := enemiesTrapped.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> enemies trapped per min\n", enemiesTrappedPerMin)
				}
				if ripTireKills, ok := heroStats.HeroSpecific["ripTireKills"]; ok {
					ripTireKillsPerMin := ripTireKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> rip tire kills per min\n", ripTireKillsPerMin)
				}
			case "lucio":
				if soundBarriersProvided, ok := heroStats.HeroSpecific["soundBarriersProvided"]; ok {
					soundBarriersProvidedPerMin := soundBarriersProvided.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> sound barriers provided per min\n", soundBarriersProvidedPerMin)
				}
			case "mccree":
				if deadeyeKills, ok := heroStats.HeroSpecific["deadeyeKills"]; ok {
					deadeyeKillsPerMin := deadeyeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> deadeye kills per min\n", deadeyeKillsPerMin)
				}
				if fanTheHammerKills, ok := heroStats.HeroSpecific["fanTheHammerKills"]; ok {
					fanTheHammerKillsPerMin := fanTheHammerKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> fan the hammer kills per min\n", fanTheHammerKillsPerMin)
				}
			case "mei":
				if damageBlocked, ok := heroStats.HeroSpecific["damageBlocked"]; ok {
					damageBlockedPerMin := damageBlocked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> damage blocked per min\n", damageBlockedPerMin)
				}
				if blizzardKills, ok := heroStats.HeroSpecific["blizzardKills"]; ok {
					blizzardKillsPerMin := blizzardKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> blizzard kills per min\n", blizzardKillsPerMin)
				}
				if enemiesFrozen, ok := heroStats.HeroSpecific["enemiesFrozen"]; ok {
					enemiesFrozenPerMin := enemiesFrozen.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> enemies frozen per min\n", enemiesFrozenPerMin)
				}
			case "mercy":
				if damageAmplified, ok := heroStats.Miscellaneous["damageAmplified"]; ok {
					damageAmplifiedPerMin := damageAmplified.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> damage amplified per min\n", damageAmplifiedPerMin)
				}
				if blasterKills, ok := heroStats.Miscellaneous["blasterKills"]; ok {
					blasterKillsPerMin := blasterKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> blaster kills per min\n", blasterKillsPerMin)
				}
				if playersResurrected, ok := heroStats.HeroSpecific["playersResurrected"]; ok {
					playersResurrectedPerMin := playersResurrected.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> players resurrected per min\n", playersResurrectedPerMin)
				}
			case "orisa":
				if damageAmplified, ok := heroStats.Miscellaneous["damageAmplified"]; ok {
					damageAmplifiedPerMin := damageAmplified.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> damage amplified per min\n", damageAmplifiedPerMin)
				}
			case "pharah":
				if barrageKills, ok := heroStats.HeroSpecific["barrageKills"]; ok {
					barrageKillsPerMin := barrageKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> barrage kills per min\n", barrageKillsPerMin)
				}
				if rocketDirectHits, ok := heroStats.HeroSpecific["rocketDirectHits"]; ok {
					rocketDirectHits := rocketDirectHits.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> rocket direct hits per min\n", rocketDirectHits)
				}
			case "reaper":
				if deathsBlossomKills, ok := heroStats.HeroSpecific["deathsBlossomKills"]; ok {
					deathsBlossomKillsPerMin := deathsBlossomKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> blossom kills per min\n", deathsBlossomKillsPerMin)
				}
				if selfHealing, ok := heroStats.Assists["selfHealing"]; ok {
					selfHealingPerMin := selfHealing.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> self healing per min\n", selfHealingPerMin)
				}
			case "reinhardt":
				if damageBlocked, ok := heroStats.HeroSpecific["damageBlocked"]; ok {
					damageBlockedPerMin := damageBlocked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> damage blocked per min\n", damageBlockedPerMin)
				}
				if chargeKills, ok := heroStats.HeroSpecific["chargeKills"]; ok {
					chargeKillsPerMin := chargeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> charge kills per min\n", chargeKillsPerMin)
				}
				if fireStrikeKills, ok := heroStats.HeroSpecific["fireStrikeKills"]; ok {
					fireStrikeKillsPerMin := fireStrikeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> fire strike kills per min\n", fireStrikeKillsPerMin)
				}
				if earthshatterKills, ok := heroStats.HeroSpecific["earthshatterKills"]; ok {
					earthshatterKillsPerMin := earthshatterKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> earthshatter kills per min\n", earthshatterKillsPerMin)
				}
			case "roadhog":
				if enemiesHooked, ok := heroStats.HeroSpecific["enemiesHooked"]; ok {
					enemiesHookedPerMin := enemiesHooked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> enemies hooked per min\n", enemiesHookedPerMin)
				}
				if wholeHogKills, ok := heroStats.HeroSpecific["wholeHogKills"]; ok {
					wholeHogKillsPerMin := wholeHogKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> whole hog kills per min\n", wholeHogKillsPerMin)
				}
				if hookAccuracy, ok := heroStats.HeroSpecific["hookAccuracy"]; ok {
					text += fmt.Sprintf("<b>%s</b> hook accuracy\n", hookAccuracy)
				}
			case "soldier76":
				if helixRocketsKills, ok := heroStats.HeroSpecific["helixRocketsKills"]; ok {
					helixRocketsKillsPerMin := helixRocketsKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> helix rockets kills per min\n", helixRocketsKillsPerMin)
				}
				if tacticalVisorKills, ok := heroStats.HeroSpecific["tacticalVisorKills"]; ok {
					tacticalVisorKillsPerMin := tacticalVisorKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> tactical visor kills per min\n", tacticalVisorKillsPerMin)
				}
				if bioticFieldHealingDone, ok := heroStats.HeroSpecific["bioticFieldHealingDone"]; ok {
					bioticFieldHealingDonePerMin := bioticFieldHealingDone.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> healing done per min\n", bioticFieldHealingDonePerMin)
				}
			case "sombra":
				if enemiesHacked, ok := heroStats.Miscellaneous["enemiesHacked"]; ok {
					enemiesHackedPerMin := enemiesHacked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> enemies hacked per min\n", enemiesHackedPerMin)
				}
				if enemiesEmpd, ok := heroStats.Miscellaneous["enemiesEmpd"]; ok {
					enemiesEmpdPerMin := enemiesEmpd.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> enemies emp'd per min\n", enemiesEmpdPerMin)
				}
			case "symmetra":
				if playersTeleported, ok := heroStats.HeroSpecific["playersTeleported"]; ok {
					playersTeleportedPerMin := playersTeleported.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> players teleported per min\n", playersTeleportedPerMin)
				}
				if sentryTurretsKills, ok := heroStats.HeroSpecific["sentryTurretsKills"]; ok {
					sentryTurretsKillsPerMin := sentryTurretsKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> sentry turrets kills per min\n", sentryTurretsKillsPerMin)
				}
			case "torbjorn":
				if torbjornKills, ok := heroStats.HeroSpecific["torbjornKills"]; ok {
					torbjornKillsPerMin := torbjornKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> torbjorn kills per min\n", torbjornKillsPerMin)
				}
				if moltenCoreKills, ok := heroStats.HeroSpecific["moltenCoreKills"]; ok {
					moltenCoreKillsPerMin := moltenCoreKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> molten core kills per min\n", moltenCoreKillsPerMin)
				}
				if turretsKills, ok := heroStats.HeroSpecific["turretsKills"]; ok {
					turretsKillsPerMin := turretsKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> turrets kills per min\n", turretsKillsPerMin)
				}
				if armorPacksCreated, ok := heroStats.HeroSpecific["armorPacksCreated"]; ok {
					armorPacksCreatedPerMin := armorPacksCreated.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> armor packs created per min\n", armorPacksCreatedPerMin)
				}
			case "tracer":
				if pulseBombsKills, ok := heroStats.HeroSpecific["pulseBombsKills"]; ok {
					pulseBombsKillsPerMin := pulseBombsKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> pulse bombs kills per min\n", pulseBombsKillsPerMin)
				}
				if pulseBombsAttached, ok := heroStats.HeroSpecific["pulseBombsAttached"]; ok {
					pulseBombsAttachedPerMin := pulseBombsAttached.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> pulse bombs attached per min\n", pulseBombsAttachedPerMin)
				}
			case "widowmaker":
				if scopedCriticalHits, ok := heroStats.HeroSpecific["scopedCriticalHits"]; ok {
					scopedCriticalHitsPerMin := scopedCriticalHits.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> scoped critical hits per min\n", scopedCriticalHitsPerMin)
				}
				if scopedAccuracy, ok := heroStats.HeroSpecific["scopedAccuracy"]; ok {
					text += fmt.Sprintf("<b>%s</b> scoped accuracy\n", scopedAccuracy)
				}
			case "winston":
				if damageBlocked, ok := heroStats.HeroSpecific["damageBlocked"]; ok {
					damageBlockedPerMin := damageBlocked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> damage blocked per min\n", damageBlockedPerMin)
				}
				if jumpPackKills, ok := heroStats.HeroSpecific["jumpPackKills"]; ok {
					jumpPackKillsPerMin := jumpPackKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> jump pack kills per min\n", jumpPackKillsPerMin)
				}
				if primalRageKills, ok := heroStats.Miscellaneous["primalRageKills"]; ok {
					primalRageKillsPerMin := primalRageKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> primal rage kills per min\n", primalRageKillsPerMin)
				}
				if playersKnockedBack, ok := heroStats.HeroSpecific["playersKnockedBack"]; ok {
					playersKnockedBackPerMin := playersKnockedBack.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> players knocked back per min\n", playersKnockedBackPerMin)
				}
			case "zarya":
				if damageBlocked, ok := heroStats.HeroSpecific["damageBlocked"]; ok {
					damageBlockedPerMin := damageBlocked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> damage blocked per min\n", damageBlockedPerMin)
				}
				if highEnergyKills, ok := heroStats.HeroSpecific["highEnergyKills"]; ok {
					highEnergyKillsPerMin := highEnergyKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> high energy kills per min\n", highEnergyKillsPerMin)
				}
				if gravitonSurgeKills, ok := heroStats.HeroSpecific["gravitonSurgeKills"]; ok {
					gravitonSurgeKillsPerMin := gravitonSurgeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> graviton surge kills per min\n", gravitonSurgeKillsPerMin)
				}
				if projectedBarriersApplied, ok := heroStats.HeroSpecific["projectedBarriersApplied"]; ok {
					projectedBarriersAppliedPerMin := projectedBarriersApplied.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> projected barriers applied per min\n", projectedBarriersAppliedPerMin)
				}
				if averageEnergy, ok := heroStats.HeroSpecific["averageEnergy"]; ok {
					text += fmt.Sprintf("<b>%0.0f%%</b> average energy\n", averageEnergy.(float64)*100)
				}
			case "zenyatta":
				if transcendenceHealing, ok := heroStats.Miscellaneous["transcendenceHealing"]; ok {
					transcendenceHealingPerMin := transcendenceHealing.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.0f</b> transcendence healing per min\n", transcendenceHealingPerMin)
				}
				if offensiveAssists, ok := heroStats.Assists["offensiveAssists"]; ok {
					offensiveAssistsPerMin := offensiveAssists.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> offensive assists per min\n", offensiveAssistsPerMin)
				}
				if defensiveAssists, ok := heroStats.Miscellaneous["defensiveAssists"]; ok {
					defensiveAssistsPerMin := defensiveAssists.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("<b>%0.2f</b> defensive assists per min\n", defensiveAssistsPerMin)
				}
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

// Fetch Overwatch profile based on region and BattleTag / PSN ID / Xbox Live Account
func GetOverwatchProfile(region string, nick string) (*ovrstat.PlayerStats, error) {
	if region == "eu" || region == "us" || region == "kr" {
		return ovrstat.PCStats(region, nick)
	} else if region == "psn" || region == "xbl" {
		return ovrstat.ConsoleStats(region, nick)
	}

	return nil, errors.New("region is wrong")
}
