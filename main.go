package main

import (
	"flag"
	cowinapi "go-cowin-bot/cowin-api"
	discordbot "go-cowin-bot/discord-bot"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	Token           string
	ChannelID       string
	ServeHTTP       = flag.Bool("http", false, "Serve HTTP")
	ServeDiscordBot = flag.Bool("discord", false, "Serve Discord Bot")
)

func init() {
	flag.Parse()
	log.Printf("serveHTTP: %v | dcordbot: %v ", *ServeHTTP, *ServeDiscordBot)
}

func main() {
	sc := make(chan os.Signal, 1)

	if *ServeHTTP == false && *ServeDiscordBot == false {
		log.Println("set flag '-http' or '-discord'")
		return
	}

	if *ServeHTTP {
		go cowinapi.Serve()
	}

	if *ServeDiscordBot {
		go discordbot.Start(sc)
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running. Press CTRL+C to exit.")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Println("Bot exiting.")
}
