package discordbot

import (
	"fmt"
	cowinapi "go-cowin-bot/cowin-api"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
)

const (
	DcordMsg = "Available Capacity: %v\nDate: %v\nMin Age: %v\nVaccine Name: %v\nFee Type: %v\nSlots: %v"
)

func Start(distID, age string, pollTimer, days int, killCh chan os.Signal) {
	Token := os.Getenv("DISCORD_TOKEN")
	if Token == "" {
		log.Println("did not find DISCORD_TOKEN in environment")
		killCh <- os.Interrupt
		return
	}

	ChannelID := os.Getenv("CHANNEL_ID")
	if ChannelID == "" {
		log.Println("did not find CHANNEL_ID in environment")
		killCh <- os.Interrupt
		return
	}

	ErrorChannel := os.Getenv("ERR_CHANNEL_ID")
	if ErrorChannel == "" {
		log.Println("did not find ERR_CHANNEL_ID in environment -- cannot report errors")
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		killCh <- os.Interrupt
		return
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		killCh <- os.Interrupt
		return
	}

	// Cleanly close down the Discord session.
	defer dg.Close()

	log.Println("Polling CoWin API:")

	maxSleep := pollTimer + 2
	minSleep := pollTimer - 2
	for {
		rand.Seed(time.Now().UnixNano())
		sleeptime := rand.Intn(maxSleep-minSleep) + minSleep

		log.Println("sleeping for: ", sleeptime, "s")

		time.Sleep(time.Duration(sleeptime) * time.Second)

		output, err := cowinapi.GetBulkAvailability(distID, age, days)
		if err != nil {
			log.Println("ERROR: ", err)
			if ErrorChannel != "" {
				dg.ChannelMessageSendEmbed(ErrorChannel, embed.NewGenericEmbedAdvanced("ERROR", err.Error(), 0x990000))
			}

			continue
		}

		if len(output) > 0 {
			dg.ChannelMessageSend(ChannelID, "NEW UPDATE:")
			for _, o := range output {
				slots := strings.Join(o.Slots, ", ")
				msg := fmt.Sprintf(DcordMsg, o.AvailableCapacity, o.Date, o.MinAge, o.VaccineName, o.FeeType, slots)
				title := fmt.Sprintf("%v - %v", o.CenterName, o.Pincode)

				dg.ChannelMessageSendEmbed(ChannelID, embed.NewGenericEmbedAdvanced(title, msg, 0xc1f175))
			}
		}
	}
}
