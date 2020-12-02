package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	queue "github.com/ethanent/discord-voicestateupdatequeue"
	"os"
	"os/signal"
)

func handleEvents(c chan *queue.VoiceStateEvent) {
	for a := range c {
		fmt.Println(a.Event, a.UserID, a.GuildID, a.ChannelID)
	}
}

func main() {
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))

	if err != nil {
		fmt.Println("Ensure you have set environment variable DISCORD_BOT_TOKEN.")

		panic(err)
	}

	c := make(chan *queue.VoiceStateEvent)

	q := queue.NewVoiceStateEventQueue(c)

	s.AddHandler(q.Handler)

	go handleEvents(c)

	if err := s.Open(); err != nil {
		panic(err)
	}

	fmt.Println("DONE")

	a := make(chan os.Signal)

	signal.Notify(a, os.Interrupt)

	<-a

	fmt.Println("Exit")

	if err := s.Close(); err != nil {
		panic(err)
	}
}
