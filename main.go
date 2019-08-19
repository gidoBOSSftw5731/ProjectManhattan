package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"

	weather "github.com/gidoBOSSftw5731/goweather"

	"github.com/bwmarrin/discordgo"
	"github.com/gidoBOSSftw5731/log"
)

var (
	botID         string
	discordToken  = flag.String("token", "", "Discord bot secret")
	apiToken      = flag.String("apikey", "", "openweathermap api key")
	commandPrefix = "!"
)

func main() {
	flag.Parse()
	log.SetCallDepth(4)
	//log.Tracef("Token is: %v", *token)
	discord, err := discordgo.New("Bot " + *discordToken)
	errCheck("error creating discord session", err)
	user, err := discord.User("@me")
	errCheck("error retrieving account", err)

	botID = user.ID
	discord.AddHandler(commandHandler)
	discord.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		err = discord.UpdateStatus(2, "Go Away!")
		if err != nil {
			log.Errorln("Error attempting to set my status")
		}
		servers := discord.State.Guilds
		log.Debugf("Weeathar has started on %d servers", len(servers))
	})

	err = discord.Open()
	errCheck("Error opening connection to Discord", err)
	defer discord.Close()

	<-make(chan struct{})

}

func errCheck(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %+v", msg, err)
	}
}

func commandHandler(discord *discordgo.Session, message *discordgo.MessageCreate) {
	user := message.Author
	if user.ID == botID || user.Bot {
		//Do nothing because the bot is talking
		return
	}

	//content := message.Content

	//fmt.Printf("Message: %+v || From: %s\n", message.Message, message.Author)

	if message.Content[:len(commandPrefix)] == commandPrefix {
		command := strings.Split(message.Content, commandPrefix)[1]
		commandContents := strings.Split(message.Content, " ") // 0 = !command, 1 = first arg, etc
		if len(commandContents) < 2 {
			log.Errorln("didnt supply both a city and state or a zipcode")
			discord.ChannelMessageSend(message.ChannelID, "Error in formatting!")
			return
		}

		switch strings.Split(command, " ")[0] {
		case "weather":
			var location string

			for i := 1; i < len(commandContents); i++ {
				location += commandContents[i] + " "
			}

			w, err := weather.CurrentWeather(location, *apiToken)
			if err != nil {
				log.Errorln("Error gathering current weather: ", err)
				discord.ChannelMessageSend(message.ChannelID, "Internal Error! \n"+fmt.Sprint(err))
				return
			}
			log.Traceln(w)

			if w.Cod == 0 {
				log.Errorln("Recieved invalid response from OWM, try again and check spelling")
				discord.ChannelMessageSend(message.ChannelID, "Recieved invalid response from OWM, try again and check spelling")
				return
			}

			embed := &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("Weather in %v (%v, %v) right now", w.Name, w.Coord.Lat, w.Coord.Lon),
				Author:      &discordgo.MessageEmbedAuthor{},
				Color:       rand.Intn(16777215), // Green
				Description: "All measurements in metric",
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name: "Conditions",
						Value: fmt.Sprintf("%v: %v \n Cloud percentage: %v",
							w.Weather[0].Main, w.Weather[0].Description, w.Clouds.All),
						Inline: true,
					},
					&discordgo.MessageEmbedField{
						Name:   "Tempuature :thermometer:",
						Value:  fmt.Sprint(w.Main.Temp),
						Inline: true,
					},
					&discordgo.MessageEmbedField{
						Name:   "Humidity",
						Value:  fmt.Sprint(w.Main.Humidity),
						Inline: true,
					},
					&discordgo.MessageEmbedField{
						Name:   "Wind :wind_blowing_face:",
						Value:  fmt.Sprintf("Speed: %v, Direction (degrees): %v", w.Wind.Speed, w.Wind.Deg),
						Inline: true,
					},
				},
				Image: &discordgo.MessageEmbedImage{
					URL: "https://cdn.discordapp.com/avatars/119249192806776836/cc32c5c3ee602e1fe252f9f595f9010e.jpg?size=2048",
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://cdn.discordapp.com/avatars/119249192806776836/cc32c5c3ee602e1fe252f9f595f9010e.jpg?size=2048",
				},
				Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
			}

			resp, err := discord.ChannelMessageSendEmbed(message.ChannelID, embed)
			if err != nil {
				log.Debugln(resp, err)
				discord.ChannelMessageSend(message.ChannelID, "Internal Error!")
				return
			}
		default:
			log.Errorln("invalid command!")
			discord.ChannelMessageSend(message.ChannelID, "Invalid command!")
			return
		}
	}

	//log.Traceln(content)
}
