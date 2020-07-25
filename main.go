package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func init() {

}

func main() {
	var configFile, logFile string

	flag.StringVar(&configFile, "c", "config.toml", "Configuration file")
	flag.StringVar(&logFile, "l", "rinako.log.txt", "Log file")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(logFile), 0770); err == nil {
		if fi, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660); err == nil {
			log.SetOutput(fi)
		} else {
			fmt.Printf("Failed to open log file due to err %v\n", err)
		}
	} else {
		fmt.Printf("Failed to create log file due to err %v\n", err)
	}

	config, err := ReadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config file %s: %v", configFile, err)
		os.Exit(1)
	}

	dg, err := discordgo.New("Bot " + config.AuthToken)

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}
