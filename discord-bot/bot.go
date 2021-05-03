package discordbot

import (
	"fmt"
	cowinapi "go-cowin-bot/cowin-api"
	"log"
	"os"
	"strings"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
)

const (
	DcordMsg = "_\nCenter Name: *%v* \nPincode: *%v*\nAvailable Capacity: %v\nDate: %v\nMin Age: %v\nVaccine Name: %v\nFee Type: %v\nSlots: %v\n----X----"
)

func Start(distID, age string, pollTimer, days int, killCh chan os.Signal) {
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

	log.Println("Polling CoWin API:")

	for {
		time.Sleep(time.Duration(pollTimer) * time.Second)

		output, err := cowinapi.GetBulkAvailability(distID, age, days)
		if err != nil {
			log.Println("ERROR: ", err)
			continue
		}

		if len(output) > 0 {
			dg.ChannelMessageSend(ChannelID, "NEW UPDATE:")
			for _, o := range output {
				slots := strings.Join(o.Slots, ", ")
				msg := fmt.Sprintf(DcordMsg, o.CenterName, o.Pincode, o.AvailableCapacity, o.Date, o.MinAge, o.VaccineName, o.FeeType, slots)
				title := fmt.Sprintf("%v - %v", o.CenterName, o.Pincode)

				dg.ChannelMessageSendEmbed(ChannelID, embed.NewGenericEmbed(title, msg))
			}
		}
	}
}
