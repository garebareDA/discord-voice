package main

import(
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
)

type state struct{
	Name string
	talking bool
	channel string
}

var(
	Token = "Bot " + os.Getenv("TOKEN")
	stopBot = make(chan bool)
	usermap = map[string]*state{}
)

func main(){
	discord, err := discordgo.New()
	discord.Token = Token
	if err != nil {
		panic(err)
	}

	discord.AddHandler(voice)

	err = discord.Open()
	if err != nil {
		panic(err)
	}

	<- stopBot
	return
}

func voice(s *discordgo.Session, vs *discordgo.VoiceStateUpdate){
	_, ok := usermap[vs.UserID]

	if !ok {
		usermap[vs.UserID] = new(state)
		user, _ := s.User(vs.UserID)
		usermap[vs.UserID].Name = user.Username
		usermap[vs.UserID].talking = false
		usermap[vs.UserID].channel = ""
	}

	if len(vs.ChannelID) > 0 && usermap[vs.UserID].talking == false{

		channel, _ := s.Channel(vs.ChannelID)
		enterning(s, vs.UserID, channel.Name)

	}else if usermap[vs.UserID].talking == true {

		message := usermap[vs.UserID].Name+"さんが"+usermap[vs.UserID].channel+"から退室しました"
		channel, _ := s.Channel(vs.ChannelID)

		channelID := channelList(s, "マンメンミ！")
		sendMessage(s, channelID, message)

		if channel != nil && channel.Type == 2{

			enterning(s, vs.UserID, channel.Name)
		}else if channel == nil{
			usermap[vs.UserID].talking = false
			usermap[vs.UserID].channel = ""
			return
		}
	}
}

func enterning(s *discordgo.Session, id string, name string){
	usermap[id].talking = true
	usermap[id].channel = name
	message := usermap[id].Name+"さんが"+name+"に入室しました"
	channelid := channelList(s, "マンメンミ！")
	sendMessage(s, channelid, message)
}

func sendMessage(s *discordgo.Session, id string, msg string) {
	_, err := s.ChannelMessageSend(id, msg)
	if err != nil {
			log.Println("Error sending message: ", err)
	}
}

func channelList(s *discordgo.Session, name string) string {
	var channnelid string

	for _, guild := range s.State.Guilds{
		channels, _ := s.GuildChannels(guild.ID)
		for _, c := range channels{
			if c.Type != discordgo.ChannelTypeGuildText {
				continue
			}
			if name == c.Name {
				channnelid = c.ID
			}
		}
	}

	return channnelid
}