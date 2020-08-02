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

var discriminator string
var allowedRoles []RoleDesc
var allowedRoleTitles []string
var roleCh string
var rinako Rinako

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

	rinako.config = config

	discriminator = rinako.config.Discriminator
	roleCh = rinako.config.RoleChannel
	allowedRoles = rinako.config.AllowedRoles
	for _, rd := range allowedRoles {
		allowedRoleTitles = append(allowedRoleTitles, rd.Role)
	}

	rinako.session, err = discordgo.New("Bot " + rinako.config.AuthToken)

	// Register the messageCreate func as a callback for MessageCreate events.
	// dg.AddHandler(roleMessageCreate)
	rinako.session.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	rinako.session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// Open a websocket connection to Discord and begin listening.
	err = rinako.session.Open()
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
	rinako.session.Close()
}

// Rinako represents an instance of Rinako
type Rinako struct {
	session *discordgo.Session
	config  *Config
}
