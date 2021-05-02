package discordbot

import (
	"fmt"
	cowinapi "go-cowin-bot/cowin-api"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	DcordMsg = "CenterID: %v\nCenterName: %v\nPincode: %v\nFeeType: %v\nAvailableCapacity: %v\nMinAge: %v\nVaccineName: %v\nSlots: %v\nDate: %v"
)

func Start(killCh chan os.Signal) {
	Token := os.Getenv("DISCORD_TOKEN")
	if Token == "" {
		log.Println("did not find DISCORD_TOKEN in environment")
		return
	}

	ChannelID := os.Getenv("CHANNEL_ID")
	if Token == "" {
		log.Println("did not find CHANNEL_ID in environment")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	// Cleanly close down the Discord session.
	defer dg.Close()

	log.Println("Polling CoWin API")

	time.Sleep(2 * time.Second)

	// todo make it poll until stopped
	for {
		time.Sleep(1 * time.Minute)

		output, err := cowinapi.HitURL("265", "18")
		if err != nil {
			log.Println("ERROR: ", err)
			dg.ChannelMessageSend(ChannelID, string(err.Error()))
			continue
		}

		if output != nil {
			for _, o := range output {
				msg := fmt.Sprintf(DcordMsg, o.CenterID, o.CenterName, o.Pincode, o.FeeType, o.AvailableCapacity, o.MinAge, o.VaccineName, o.Slots, o.Date)
				dg.ChannelMessageSend(ChannelID, string(msg))
			}
		}
	}
}
