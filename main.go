package main

import (
	"flag"
	"regexp"
	"strings"
	"time"

	"github.com/gidoBOSSftw5731/yahooweather"

	"github.com/bwmarrin/discordgo"
	"github.com/gidoBOSSftw5731/log"
)

var (
	botID         string
	token         = flag.String("token", "", "Discord bot secret")
	commandPrefix = "!"
)

func main() {
	flag.Parse()
	log.SetCallDepth(4)
	//log.Tracef("Token is: %v", *token)
	discord, err := discordgo.New("Bot " + *token)
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

	if !strings.HasPrefix(commandPrefix, message.Content) {
		command := strings.Split(message.Content, commandPrefix)[1]
		commandContents := strings.Split(message.Content, " ") // 0 = !command, 1 = first arg, etc
		if len(commandContents) < 2 {
			log.Errorln("didnt supply both a city and state or a zipcode")
			discord.ChannelMessageSend(message.ChannelID, "Error in formatting!")
			return
		}

		switch strings.Split(command, " ")[0] {
		case "weather":

			isZipcode, err := regexp.MatchString("^\\d{5}(?:[\\s]\\d{4})?$",
				commandContents[1])
			if err != nil {
				log.Errorln(err)
				discord.ChannelMessageSend(message.ChannelID, "Error in Regex!")
				return
			}

			var loc *yahooweather.Location

			if isZipcode {
				loc = yahooweather.BuildLocation("", "", commandContents[1])
			} else {
				cleanString := strings.TrimPrefix(message.Content,
					commandPrefix+command+" ")
				cleanString = strings.Replace(cleanString, ",", "", 0)
				split := strings.Split(cleanString, " ")

				if len(split) < 2 {
					log.Errorln("didnt supply both a city and state")
					discord.ChannelMessageSend(message.ChannelID, "Error in formatting!")
					return
				}

				loc = yahooweather.BuildLocation(split[0], split[1], "")
			}

			url := yahooweather.BuildUrl(loc)

			log.Traceln(url)

			weather := yahooweather.MakeQuery(url)

			log.Debugln(weather)

			embed := &discordgo.MessageEmbed{
				Author:      &discordgo.MessageEmbedAuthor{},
				Color:       0x00ff00, // Green
				Description: "This is a discordgo embed",
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:   "tempuature",
						Value:  weather.Temp,
						Inline: true,
					},
					&discordgo.MessageEmbedField{
						Name:   "Weth",
						Value:  weather.Weth,
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
				Title:     "I am an Embed",
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
