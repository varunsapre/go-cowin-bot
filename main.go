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
	Token      string
	ChannelID  string
	districtID string
	age        string
	pollTimer  int
	days       int

	ServeHTTP       = flag.Bool("http", false, "Serve HTTP")
	ServeDiscordBot = flag.Bool("discord", false, "Serve Discord Bot")
)

func init() {
	flag.StringVar(&districtID, "district_id", "294", "district id for bot to check")
	flag.StringVar(&age, "age", "18", "minimum age")

	flag.IntVar(&pollTimer, "poll", 15, "number of seconds for polling")
	flag.IntVar(&days, "days", 10, "number of days to check ahead")

	flag.Parse()

	log.Printf("serveHTTP: %v | dcordbot: %v ", *ServeHTTP, *ServeDiscordBot)
	log.Printf("distID: %v | minAge: %v | pollTimer: %v | days: %v", districtID, age, pollTimer, days)
}

func main() {
	sc := make(chan os.Signal, 1)

	if !*ServeHTTP && !*ServeDiscordBot {
		log.Println("set flag '-http' or '-discord'")
		return
	}

	if *ServeHTTP {
		go cowinapi.Serve()
	}

	if *ServeDiscordBot {
		go discordbot.Start(districtID, age, pollTimer, days, sc)
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running. Press CTRL+C to exit.")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Bot exiting.")
}
