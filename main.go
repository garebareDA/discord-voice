package main

import(
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
	"strings"
)

type state struct{
	Name string
	talking bool
	channel string
	channelID string
}

var(
	Token = "Bot " + os.Getenv("TOKEN")
	stopBot = make(chan bool)
	usermap = map[string]*state{}
	notifiedChannel = map[string]string{}
)

func main(){
	discord, err := discordgo.New()
	discord.Token = Token
	if err != nil {
		panic(err)
	}

	discord.AddHandler(voice)
	discord.AddHandler(messageCatch)

	err = discord.Open()
	if err != nil {
		panic(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func voice(s *discordgo.Session, vs *discordgo.VoiceStateUpdate){
	_, ok := usermap[vs.UserID]

	if !ok {
		usermap[vs.UserID] = new(state)
		user, _ := s.User(vs.UserID)
		usermap[vs.UserID].Name = user.Username
		usermap[vs.UserID].talking = false
		usermap[vs.UserID].channel = ""
		usermap[vs.UserID].channelID = ""
	}

	if len(vs.ChannelID) > 0 && usermap[vs.UserID].talking == false{

		channel, _ := s.Channel(vs.ChannelID)
		enterning(s, vs.UserID, channel.Name, channel.GuildID, notifiedChannel[channel.GuildID])

	}else if usermap[vs.UserID].talking == true {

		message := usermap[vs.UserID].Name+"さんが"+usermap[vs.UserID].channel+"から退室しました"
		channel, _ := s.Channel(vs.ChannelID)
		channelid := usermap[vs.UserID].channelID
		if len(channelid) > 0 {
			sendMessage(s, channelid, message)
		}

		if channel != nil && channel.Type == 2{

			enterning(s, vs.UserID, channel.Name, channel.GuildID, notifiedChannel[channel.GuildID])
		}else if channel == nil{
			usermap[vs.UserID].talking = false
			usermap[vs.UserID].channel = ""
		}
	}
}

func enterning(s *discordgo.Session, id string, name string, guildID string, channelID string){
	usermap[id].talking = true
	usermap[id].channel = name
	usermap[id].channelID = channelID
	message := usermap[id].Name+"さんが"+name+"に入室しました"
	channelid := notifiedChannel[guildID]
	if len(channelid) > 0{
		sendMessage(s, channelid, message)
	}
}

func sendMessage(s *discordgo.Session, id string, msg string) {
	_, err := s.ChannelMessageSend(id, msg)
	if err != nil {
			log.Println("Error sending message: ", err)
	}
}

func channelList(s *discordgo.Session, name string, guildID string) bool {
	channelIs := false

	for _, guild := range s.State.Guilds{
		channels, _ := s.GuildChannels(guild.ID)
		for _, c := range channels{
			if c.Type != discordgo.ChannelTypeGuildText {
				continue
			}
			if name == c.Name && guildID == c.GuildID{
				notifiedChannel[c.GuildID] = c.ID
				channelIs = true
			}
		}
	}
	return channelIs
}

func messageCatch(s *discordgo.Session, m *discordgo.MessageCreate){
	if m.Author.ID == s.State.User.ID {
		return
	}

	commands := strings.Split(m.Content, " ")
	command := commands[0]

	if command == "/noti" {
		channelIs := commands[1]
		name := channelList(s, channelIs, m.GuildID)
		if name == true {
			s.ChannelMessageSend(m.ChannelID, channelIs + "に設定しました")
		}else{
			s.ChannelMessageSend(m.ChannelID, "そのチャンネルは存在していません")
		}
	}
}