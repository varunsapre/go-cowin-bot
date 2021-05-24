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
	Token     string
	ChannelID string
	ops       cowinapi.Options

	Cmd             = flag.Bool("cmd", false, "Serve HTTP")
	ServeDiscordBot = flag.Bool("discord", false, "Serve Discord Bot")
)

func init() {
	ops = cowinapi.Options{}
	flag.StringVar(&ops.DistrictID, "district_id", "294", "district id for bot to check")
	flag.StringVar(&ops.VaccineName, "vaccine", "", "vaccine name filter")

	flag.IntVar(&ops.Age, "age", 18, "minimum age")
	flag.IntVar(&ops.AvailableCapacity, "minCapacity", 2, "minimum capacity")
	flag.IntVar(&ops.PollTimer, "poll", 15, "number of seconds for polling")
	flag.IntVar(&ops.Days, "days", 10, "number of days to check ahead")
	flag.IntVar(&ops.DoseNum, "doseNumber", 1, "1 or 2")

	flag.BoolVar(&ops.RunAtNight, "night", false, "Serve Discord Bot")

	flag.Parse()

	log.Printf("Options:\n\t%#v", ops)
}

func main() {
	sc := make(chan os.Signal, 1)

	if !*Cmd && !*ServeDiscordBot {
		log.Println("set flag '-cmd' or '-discord'")
		return
	}

	if *Cmd {
		go cowinapi.StartCMDOnly(&ops)

		// force discord bot to not start
		*ServeDiscordBot = false
	}

	if *ServeDiscordBot {
		go discordbot.Start(&ops, sc)
	}

	log.Printf("serveHTTP: %v | dcordbot: %v ", *Cmd, *ServeDiscordBot)

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running. Press CTRL+C to exit.")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Bot exiting.")
}
