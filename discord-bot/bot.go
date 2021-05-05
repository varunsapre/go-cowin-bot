package discordbot

import (
	"errors"
	"fmt"
	cowinapi "go-cowin-bot/cowin-api"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	DcordMsg = "Available Capacity: %v\nDate: %v\nMin Age: %v\nVaccine Name: %v\nFee Type: %v\nSlots: %v"
)

var (
	Token        string
	ChannelID    string
	GuildID      string
	ErrChannelId string
)

func setEnvFromFile() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func readOsEnv() error {
	Token := os.Getenv("DISCORD_TOKEN")
	if Token == "" {
		return errors.New("missing DISCORD_TOKEN in env")
	}

	ChannelID := os.Getenv("CHANNEL_ID")
	if ChannelID == "" {
		return errors.New("missing CHANNEL_ID in env")
	}

	GuildID := os.Getenv("GUILD_ID")
	if GuildID == "" {
		return errors.New("missing GUILD_ID in env")
	}

	ErrChannelId := os.Getenv("ERR_CHANNEL_ID")
	if ErrChannelId == "" {
		log.Println("did not find ERR_CHANNEL_ID in environment -- cannot report errors")
	}

	return nil
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name: "basic-command",
		// All commands and options must have a description
		// Commands/options without description will fail the registration
		// of the command.
		Description: "Basic command",
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"basic-command": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: "Hey there! Congratulations, you just executed your first slash command",
			},
		})
	},
}

func Start(distID, age string, pollTimer, days int, killCh chan os.Signal) {
	setEnvFromFile()
	readOsEnv()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		killCh <- os.Interrupt
		return
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.Data.Name]; ok {
			h(s, i)
		}
	})

	for _, v := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
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
			if ErrChannelId != "" {
				dg.ChannelMessageSendEmbed(ErrChannelId, embed.NewGenericEmbedAdvanced("ERROR", err.Error(), 0x990000))
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
