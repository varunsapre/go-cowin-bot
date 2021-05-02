package discordbot

import (
	"fmt"
	cowinapi "go-cowin-bot/cowin-api"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	DcordMsg = " \nCenter Name: %v\nAvailable Capacity: %v\nMin Age: %v\nVaccine Name: %v\nFee Type: %v\nSlots: %v\nDate: %v\nPincode: %v"
)

func Start(killCh chan os.Signal) {
	Token := os.Getenv("DISCORD_TOKEN")
	if Token == "" {
		log.Println("did not find DISCORD_TOKEN in environment")
		killCh <- os.Kill
		return
	}

	ChannelID := os.Getenv("CHANNEL_ID")
	if ChannelID == "" {
		log.Println("did not find CHANNEL_ID in environment")
		killCh <- os.Kill
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		killCh <- os.Kill
		return
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		killCh <- os.Kill
		return
	}

	// Cleanly close down the Discord session.
	defer dg.Close()

	log.Println("Polling CoWin API")

	// todo make it poll until stopped
	for {
		time.Sleep(5 * time.Minute)

		output, err := cowinapi.HitURL("265", "18")
		if err != nil {
			log.Println("ERROR: ", err)
			continue
		}

		if output != nil {
			dg.ChannelMessageSend(ChannelID, "------------------------------------------------------------------------------")
			for _, o := range output {
				slots := strings.Join(o.Slots, ", ")
				msg := fmt.Sprintf(DcordMsg, o.CenterName, o.AvailableCapacity, o.MinAge, o.VaccineName, o.FeeType, slots, o.Date, o.Pincode)

				dg.ChannelMessageSend(ChannelID, string(msg))
			}
		}
	}
}
