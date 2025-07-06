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

func main() {
	rand.Seed(time.Now().UnixNano())

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

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
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", cmd.Name, err)
		}
	}
	log.Println("Slash commands registered successfully!")
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := i.ApplicationCommandData().Name
	
	switch commandName {
	case "help":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "qwerty123",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	case "pn", "patchnotes":
		patchNotesURL := getLatestPatchNotesURL()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("📋 **Latest League of Legends Patch Notes:**\n%s", patchNotesURL),
			},
		})
	}
}
