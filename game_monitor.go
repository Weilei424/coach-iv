package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

type GameMonitor struct {
	db       *Database
	riotAPI  *RiotAPI
	discord  *discordgo.Session
	cron     *cron.Cron
	channelID string
}

func NewGameMonitor(db *Database, riotAPI *RiotAPI, discord *discordgo.Session, channelID string) *GameMonitor {
	return &GameMonitor{
		db:        db,
		riotAPI:   riotAPI,
		discord:   discord,
		cron:      cron.New(),
		channelID: channelID,
	}
}

func (gm *GameMonitor) Start() {
	gm.cron.AddFunc("@every 5m", gm.checkForNewGames)
	gm.cron.Start()
	log.Println("Game monitor started - checking for new games every 5 minutes")
}

func (gm *GameMonitor) Stop() {
	gm.cron.Stop()
	log.Println("Game monitor stopped")
}

func (gm *GameMonitor) checkForNewGames() {
	players, err := gm.db.GetTrackedPlayers()
	if err != nil {
		log.Printf("Error getting tracked players: %v", err)
		return
	}

	for _, player := range players {
		if err := gm.checkPlayerForNewGames(player); err != nil {
			log.Printf("Error checking games for %s#%s: %v", player.GameName, player.TagLine, err)
		}
	}
}

func (gm *GameMonitor) checkPlayerForNewGames(player TrackedPlayer) error {
	matchIDs, err := gm.riotAPI.GetMatchHistory(player.PUUID, 5)
	if err != nil {
		return err
	}

	if len(matchIDs) == 0 {
		return nil
	}

	latestMatchID := matchIDs[0]
	if latestMatchID == player.LastMatchID {
		return nil
	}

	for _, matchID := range matchIDs {
		if matchID == player.LastMatchID {
			break
		}

		if err := gm.processNewMatch(player, matchID); err != nil {
			log.Printf("Error processing match %s: %v", matchID, err)
			continue
		}
	}

	return gm.db.UpdateLastMatchID(player.PUUID, latestMatchID)
}

func (gm *GameMonitor) processNewMatch(player TrackedPlayer, matchID string) error {
	match, err := gm.riotAPI.GetMatchDetails(matchID)
	if err != nil {
		return err
	}

	matchData := gm.riotAPI.ExtractPlayerData(match, player.PUUID)
	if matchData == nil {
		return fmt.Errorf("player not found in match data")
	}

	if err := gm.db.AddMatchData(matchData); err != nil {
		return err
	}

	gm.sendGameSummary(player, matchData)
	return nil
}

func (gm *GameMonitor) sendGameSummary(player TrackedPlayer, match *MatchData) {
	if gm.channelID == "" {
		return
	}

	winStatus := "🔴 Loss"
	if match.Win {
		winStatus = "🟢 Win"
	}

	kda := fmt.Sprintf("%.2f", float64(match.Kills+match.Assists)/float64(max(match.Deaths, 1)))
	
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🎮 New Game Detected - %s#%s", player.GameName, player.TagLine),
		Color: func() int {
			if match.Win {
				return 0x00FF00
			}
			return 0xFF0000
		}(),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Result",
				Value:  winStatus,
				Inline: true,
			},
			{
				Name:   "Champion",
				Value:  match.Champion,
				Inline: true,
			},
			{
				Name:   "KDA",
				Value:  fmt.Sprintf("%d/%d/%d (%.2s)", match.Kills, match.Deaths, match.Assists, kda),
				Inline: true,
			},
			{
				Name:   "CS",
				Value:  fmt.Sprintf("%d", match.CreepScore),
				Inline: true,
			},
			{
				Name:   "Damage",
				Value:  fmt.Sprintf("%,d", match.DamageDealt),
				Inline: true,
			},
			{
				Name:   "Vision Score",
				Value:  fmt.Sprintf("%d", match.VisionScore),
				Inline: true,
			},
			{
				Name:   "Game Mode",
				Value:  strings.Title(strings.ReplaceAll(match.GameMode, "_", " ")),
				Inline: true,
			},
			{
				Name:   "Duration",
				Value:  fmt.Sprintf("%d:%02d", match.GameDuration/60, match.GameDuration%60),
				Inline: true,
			},
			{
				Name:   "Gold Earned",
				Value:  fmt.Sprintf("%,d", match.GoldEarned),
				Inline: true,
			},
		},
		Timestamp: match.GameCreation.Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Match ID: %s", match.MatchID),
		},
	}

	_, err := gm.discord.ChannelMessageSendEmbed(gm.channelID, embed)
	if err != nil {
		log.Printf("Error sending game summary: %v", err)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}