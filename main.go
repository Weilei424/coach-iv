package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	db          *Database
	riotAPI     *RiotAPI
	gameMonitor *GameMonitor
)

func main() {
	rand.Seed(time.Now().UnixNano())

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	riotAPIKey := os.Getenv("RIOT_API_KEY")
	if riotAPIKey == "" {
		log.Fatal("RIOT_API_KEY environment variable is required")
	}

	// Initialize database
	var err error
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "lol_bot"
	}

	db, err = NewDatabase(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.Close()

	// Initialize Riot API client
	riotAPI = NewRiotAPI(riotAPIKey)

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(interactionCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening connection:", err)
	}

	registerSlashCommands(dg)

	// Also register guild-specific commands for faster testing
	guilds := dg.State.Guilds
	for _, guild := range guilds {
		registerGuildSlashCommands(dg, guild.ID)
	}

	// Initialize and start game monitor
	monitorChannelID := os.Getenv("MONITOR_CHANNEL_ID")
	gameMonitor = NewGameMonitor(db, riotAPI, dg, monitorChannelID)
	gameMonitor.Start()
	defer gameMonitor.Stop()

	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func getLatestPatchNotesURL() string {
	// Try to get the latest patch notes URL
	resp, err := http.Get("https://www.leagueoflegends.com/en-us/news/tags/patch-notes/")
	if err != nil {
		log.Printf("Error fetching patch notes page: %v", err)
		return "https://www.leagueoflegends.com/en-us/news/tags/patch-notes/"
	}
	defer resp.Body.Close()

	// For now, return the main patch notes page
	// This could be enhanced to parse the HTML and extract the latest patch URL
	return "https://www.leagueoflegends.com/en-us/news/tags/patch-notes/"
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.Contains(strings.ToLower(m.Content), "<@"+s.State.User.ID+">") {
		responses := []string{
			"Hi👋 你带砍倒了吗?",
			"稍等 我去瞅眼Patch notes",
			"今天的比赛你看了吗",
			"这就是慢性死亡...",
		}

		randomResponse := responses[rand.Intn(len(responses))]
		s.ChannelMessageSend(m.ChannelID, randomResponse)
	}
}

func registerGuildSlashCommands(s *discordgo.Session, guildID string) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Show help information",
		},
		{
			Name:        "pn",
			Description: "Get the latest League of Legends patch notes",
		},
		{
			Name:        "patchnotes",
			Description: "Get the latest League of Legends patch notes",
		},
		{
			Name:        "track",
			Description: "Track a League of Legends player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "summoner",
					Description: "Summoner name (e.g., PlayerName#TAG)",
					Required:    true,
				},
			},
		},
		{
			Name:        "untrack",
			Description: "Stop tracking a League of Legends player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "summoner",
					Description: "Summoner name (e.g., PlayerName#TAG)",
					Required:    true,
				},
			},
		},
		{
			Name:        "stats",
			Description: "Show stats for a tracked player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "summoner",
					Description: "Summoner name (e.g., PlayerName#TAG)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "days",
					Description: "Number of days to look back (default: 7)",
					Required:    false,
				},
			},
		},
		{
			Name:        "tracked",
			Description: "List all tracked players",
		},
	}

	log.Printf("Registering %d guild-specific slash commands for guild %s...", len(commands), guildID)
	for _, cmd := range commands {
		createdCmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			log.Printf("Cannot create guild command '%v': %v", cmd.Name, err)
		} else {
			log.Printf("Successfully created guild command: %s (ID: %s)", createdCmd.Name, createdCmd.ID)
		}
	}
}

func registerSlashCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Show help information",
		},
		{
			Name:        "pn",
			Description: "Get the latest League of Legends patch notes",
		},
		{
			Name:        "patchnotes",
			Description: "Get the latest League of Legends patch notes",
		},
		{
			Name:        "track",
			Description: "Track a League of Legends player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "summoner",
					Description: "Summoner name (e.g., PlayerName#TAG)",
					Required:    true,
				},
			},
		},
		{
			Name:        "untrack",
			Description: "Stop tracking a League of Legends player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "summoner",
					Description: "Summoner name (e.g., PlayerName#TAG)",
					Required:    true,
				},
			},
		},
		{
			Name:        "stats",
			Description: "Show stats for a tracked player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "summoner",
					Description: "Summoner name (e.g., PlayerName#TAG)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "days",
					Description: "Number of days to look back (default: 7)",
					Required:    false,
				},
			},
		},
		{
			Name:        "tracked",
			Description: "List all tracked players",
		},
	}

	log.Printf("Registering %d slash commands...", len(commands))
	for _, cmd := range commands {
		createdCmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", cmd.Name, err)
		} else {
			log.Printf("Successfully created command: %s (ID: %s)", createdCmd.Name, createdCmd.ID)
		}
	}
	log.Println("Slash commands registration completed!")
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Printf("Interaction received: %s", i.ApplicationCommandData().Name)

	commandName := i.ApplicationCommandData().Name

	switch commandName {
	case "help":
		helpText := `**League of Legends Bot Commands:**

🎮 **Player Tracking:**
• /track <summoner> - Track a player's games (e.g., /track PlayerName#TAG)
• /untrack <summoner> - Stop tracking a player
• /stats <summoner> [days] - Show player stats (default: 7 days)
• /tracked - List all tracked players

📋 **Other Commands:**
• /pn or /patchnotes - Get latest patch notes

The bot will automatically post game summaries when tracked players finish games!`

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: helpText,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	case "pn", "patchnotes":
		log.Printf("Processing patch notes command: %s", commandName)
		patchNotesURL := getLatestPatchNotesURL()
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("📋 **Latest League of Legends Patch Notes:**\n%s", patchNotesURL),
			},
		})
		if err != nil {
			log.Printf("Error responding to interaction: %v", err)
		}
	case "track":
		handleTrackCommand(s, i)
	case "untrack":
		handleUntrackCommand(s, i)
	case "stats":
		handleStatsCommand(s, i)
	case "tracked":
		handleTrackedCommand(s, i)
	}
}

func handleTrackCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	summonerName := i.ApplicationCommandData().Options[0].StringValue()
	parts := strings.Split(summonerName, "#")
	if len(parts) != 2 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Invalid format. Please use: PlayerName#TAG",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	gameName, tagLine := parts[0], parts[1]

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🔍 Looking up player %s#%s...", gameName, tagLine),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	account, err := riotAPI.GetAccountByRiotID(gameName, tagLine)
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("❌ Error finding player %s#%s: %v", gameName, tagLine, err),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	summoner, err := riotAPI.GetSummonerByPUUID(account.PUUID)
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("❌ Error getting summoner data: %v", err),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	matchIDs, err := riotAPI.GetMatchHistory(account.PUUID, 1)
	if err != nil {
		log.Printf("Warning: Could not get match history for initial setup: %v", err)
	}

	var lastMatchID string
	if len(matchIDs) > 0 {
		lastMatchID = matchIDs[0]
	}

	player := &TrackedPlayer{
		PUUID:       account.PUUID,
		GameName:    gameName,
		TagLine:     tagLine,
		SummonerID:  summoner.ID,
		LastMatchID: lastMatchID,
	}

	if err := db.AddTrackedPlayer(player); err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("❌ Error adding player to database: %v", err),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("✅ Now tracking %s#%s (Level %d)", gameName, tagLine, summoner.SummonerLevel),
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

func handleUntrackCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	summonerName := i.ApplicationCommandData().Options[0].StringValue()
	parts := strings.Split(summonerName, "#")
	if len(parts) != 2 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Invalid format. Please use: PlayerName#TAG",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	gameName, tagLine := parts[0], parts[1]

	player, err := db.GetPlayerByRiotID(gameName, tagLine)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Player %s#%s is not being tracked", gameName, tagLine),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if err := db.RemoveTrackedPlayer(player.PUUID); err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Error removing player: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ Stopped tracking %s#%s", gameName, tagLine),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleStatsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	summonerName := i.ApplicationCommandData().Options[0].StringValue()
	parts := strings.Split(summonerName, "#")
	if len(parts) != 2 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Invalid format. Please use: PlayerName#TAG",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	days := 7
	if len(i.ApplicationCommandData().Options) > 1 {
		days = int(i.ApplicationCommandData().Options[1].IntValue())
	}

	gameName, tagLine := parts[0], parts[1]

	player, err := db.GetPlayerByRiotID(gameName, tagLine)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Player %s#%s is not being tracked", gameName, tagLine),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	matches, err := db.GetPlayerStats(player.PUUID, days)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Error getting stats: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if len(matches) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("📊 No games found for %s#%s in the last %d days", gameName, tagLine, days),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	wins := 0
	totalKills := 0
	totalDeaths := 0
	totalAssists := 0
	totalCS := 0
	totalDamage := 0

	for _, match := range matches {
		if match.Win {
			wins++
		}
		totalKills += match.Kills
		totalDeaths += match.Deaths
		totalAssists += match.Assists
		totalCS += match.CreepScore
		totalDamage += match.DamageDealt
	}

	winRate := float64(wins) / float64(len(matches)) * 100
	avgKDA := float64(totalKills+totalAssists) / float64(max(totalDeaths, 1))

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("📊 Stats for %s#%s (Last %d days)", gameName, tagLine, days),
		Color: 0x0099FF,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Games Played",
				Value:  fmt.Sprintf("%d", len(matches)),
				Inline: true,
			},
			{
				Name:   "Win Rate",
				Value:  fmt.Sprintf("%.1f%% (%d wins)", winRate, wins),
				Inline: true,
			},
			{
				Name:   "Average KDA",
				Value:  fmt.Sprintf("%.1f/%.1f/%.1f (%.2f)", float64(totalKills)/float64(len(matches)), float64(totalDeaths)/float64(len(matches)), float64(totalAssists)/float64(len(matches)), avgKDA),
				Inline: true,
			},
			{
				Name:   "Average CS",
				Value:  fmt.Sprintf("%.1f", float64(totalCS)/float64(len(matches))),
				Inline: true,
			},
			{
				Name:   "Average Damage",
				Value:  fmt.Sprintf("%,.0f", float64(totalDamage)/float64(len(matches))),
				Inline: true,
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleTrackedCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	players, err := db.GetTrackedPlayers()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Error getting tracked players: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if len(players) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "📋 No players are currently being tracked",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	var content strings.Builder
	content.WriteString("📋 **Currently Tracked Players:**\n\n")
	for _, player := range players {
		content.WriteString(fmt.Sprintf("• %s#%s (Added: %s)\n", player.GameName, player.TagLine, player.CreatedAt.Format("2006-01-02")))
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content.String(),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
