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
	userID := vs.UserID
	_, ok := usermap[userID]

	if !ok {
		usermap[userID] = new(state)
		user, _ := s.User(userID)
		usermap[userID].Name = user.Username
		usermap[userID].talking = false
		usermap[userID].channel = ""
		usermap[userID].channelID = ""
	}

	if len(vs.ChannelID) > 0 && usermap[userID].talking == false{

		channel, _ := s.Channel(vs.ChannelID)
		enterning(s, userID, channel.Name, notifiedChannel[channel.GuildID])

	}else if usermap[userID].talking == true {

		message := usermap[userID].Name+"さんが"+usermap[userID].channel+"から退室しました"
		channel, _ := s.Channel(vs.ChannelID)
		channelid := usermap[userID].channelID
		if len(channelid) > 0 {
			sendMessage(s, channelid, message)
		}

		if channel != nil && channel.Type == 2{

			enterning(s, userID, channel.Name, notifiedChannel[channel.GuildID])
		}else if channel == nil{
			usermap[userID].talking = false
			usermap[userID].channel = ""
		}
	}
}

func enterning(s *discordgo.Session, id string, name string, channelID string){
	usermap[id].talking = true
	usermap[id].channel = name
	usermap[id].channelID = channelID
	message := usermap[id].Name+"さんが"+name+"に入室しました"
	if len(channelID) > 0{
		sendMessage(s, channelID, message)
	}
}

func sendMessage(s *discordgo.Session, id string, msg string) {
	_, err := s.ChannelMessageSend(id, msg)
	if err != nil {
			log.Println("Error sending message: ", err)
	}
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
		if len(name) > 0 {
			notifiedChannel[m.GuildID] = name
			s.ChannelMessageSend(m.ChannelID, channelIs + "に設定しました")
		}else{
			s.ChannelMessageSend(m.ChannelID, "そのチャンネルは存在していません")
		}
	}
}

func channelList(s *discordgo.Session, name string, guildID string) string{
	id := ""

	for _, guild := range s.State.Guilds{
		channels, _ := s.GuildChannels(guild.ID)
		for _, c := range channels{
			if c.Type != discordgo.ChannelTypeGuildText {
				continue
			}
			if name == c.Name && guildID == c.GuildID{
				id = c.ID
			}
		}
	}
	return id
}