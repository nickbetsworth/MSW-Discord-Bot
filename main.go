package main

import (
	"strings"
	"flag"
  "fmt"
  "os"
  "os/signal"
  "syscall"
  "github.com/bwmarrin/discordgo"
  "github.com/nickbetsworth/mswclient"
)

func init() {
  flag.StringVar(&token, "t", "", "Bot Token")
  flag.Parse()
}

var token string

func main() {
  if token == "" {
    fmt.Println("No token was provided. Please run: mswdiscordbot -t <bot token>")
    return
  }

  spots := mswclient.GetSpots("Porthcawl")
  fmt.Println(spots[0].ID)

  forecast := mswclient.GetForecast(spots[0].ID)
  fmt.Printf("Solid rating: %d\n", forecast[0].SolidRating)
  fmt.Printf("Faded rating: %d\n", forecast[0].FadedRating)

  discord, err := discordgo.New("Bot " + token)

  if err != nil {
    fmt.Println("Error initialising discord session: ", err)
  }
  
  discordgo.MakeIntent(discordgo.IntentsGuilds | discordgo.IntentsGuildMessages)

  discord.AddHandler(messageCreate)

  err = discord.Open()

  if err != nil {
    fmt.Println("Error opening discord session: ", err)
  }

  fmt.Println("MSW forecast bot is now running. press Ctrl-C to exit.")
  sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
  <-sc
  
  discord.Close()
}

func messageCreate(s *discordgo.Session, event *discordgo.MessageCreate) {
    // Ignore messages created by the bot and non-msw commands
    if s.State.User.ID == event.Author.ID || !strings.HasPrefix(event.Content, "!msw") {
      return
    }

    searchString := strings.TrimSpace(strings.TrimPrefix(event.Content, "!msw"))

    fmt.Println(searchString)

    // spots := mswclient.GetSpots(searchString)
    // for 
    s.ChannelMessageSend(event.ChannelID, "Searching for surf spots at " + searchString)
}