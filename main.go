package main

import (
	"strings"
	"flag"
  "fmt"
  "os"
  "os/signal"
  "syscall"
  "time"
  "github.com/bwmarrin/discordgo"
  "github.com/nickbetsworth/mswclient"
)

func init() {
  flag.StringVar(&token, "t", "", "Bot Token")
  flag.Parse()
}

var token string

var emojiIds = map[string]string{
  ":SolidStar:": "764443959409770496",
  ":FadedStar:": "764443958993485855",
  ":NoStar:": "764443959157194773",
}

func main() {
  if token == "" {
    fmt.Println("No token was provided. Please run: mswdiscordbot -t <bot token>")
    return
  }

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
    s.ChannelMessageSend(event.ChannelID, "Searching for surf spots at " + searchString)
    
    // Todo: add support for direct querying by spot ID

    spots := mswclient.GetSpots(searchString)
    
    if len(spots) == 0 {
      s.ChannelMessageSend(event.ChannelID, "No spots found matching \"" + searchString + "\"")
      return
    }

    // Todo: remove output of spot list
    // Todo: if there are multiple spots, send the user a message notifying them to be more specific
    for _, spot := range spots {
      s.ChannelMessageSend(event.ChannelID, spot.Name)
    }

    spot := spots[0]
    forecast := mswclient.GetForecast(spot.ID)

    groupedForecasts := groupForecastsByDay(forecast)

    for _, dayForecast := range groupedForecasts {
      tides := mswclient.GetTides(spot.ID, dayForecast.ForecastStartTimestamp, dayForecast.ForecastStartTimestamp)

      if len(tides) > 1 || len(tides) == 0 {
        fmt.Printf("Unexpected number of tide results (%d) for spot %d", len(tides), spot.ID)
        continue
      }

      tideMessage := convertTideToMessage(tides[0].TideResult)
      forecastMessage := convertDayForecastToMessage(dayForecast)

      startOfDay := time.Unix(dayForecast.ForecastStartTimestamp, 0)

      _, err := s.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
        Title: startOfDay.Format("Monday 02/01"),
        Fields: []*discordgo.MessageEmbedField{&tideMessage, &forecastMessage},
      })

      if (err != nil) {
        fmt.Println(err)
      }
    }
}

func convertDayForecastToMessage(forecast dayForecast) discordgo.MessageEmbedField {
  result := discordgo.MessageEmbedField{Name: "Surf forecast"}

  for _, f := range forecast.ForecastPeriods {
    result.Value += convertForecastPeriodToString(f)
  }

  return result;
}

func convertTideToMessage(tides mswclient.TideResult) discordgo.MessageEmbedField {
  var result discordgo.MessageEmbedField
  result.Name = "Tides"

  var tideStrings []string

  for _, t := range tides.Tide {
    tm := time.Unix(t.Timestamp, 0)
    tideStrings = append(tideStrings, fmt.Sprintf("%s %s", t.State, tm.Format("3:04pm")))
  }

  result.Value = strings.Join(tideStrings, " | ")

  return result
}

func convertForecastPeriodToString(f mswclient.ForecastResult) string {
  time := time.Unix(f.Timestamp, 0)

  formattedHour := time.Format("3pm")
  return fmt.Sprintf(
    "%s %d-%d%s %s%.1f%s %ds | %d%s %s\n",
    formattedHour,
    f.Swell.MinBreakingHeight, 
    f.Swell.MaxBreakingHeight, 
    f.Swell.Unit, 
    getStarRatingString(f), 
    f.Swell.Height, 
    f.Swell.Unit, 
    f.Swell.Period,
    f.Wind.Speed,
    f.Wind.Unit,
    f.Wind.StringDirection,
  )
}

func groupForecastsByDay(ungroupedForecasts mswclient.ForecastResults) []dayForecast {
  var groupedForecasts []dayForecast
  
  currentDay := -1

  for _, forecastPeriod := range ungroupedForecasts {
    tm := time.Unix(forecastPeriod.Timestamp, 0)

    if tm.Day() != currentDay {
      groupedForecasts = append(groupedForecasts, dayForecast{ForecastStartTimestamp: forecastPeriod.Timestamp})
      currentDay = tm.Day()
    }

    currentDayForecast := &groupedForecasts[len(groupedForecasts)-1]
    currentDayForecast.ForecastPeriods = append(currentDayForecast.ForecastPeriods, forecastPeriod.ForecastResult)
  }

  return groupedForecasts
}

func getStarRatingString(f mswclient.ForecastResult) string {
  var output string

  for i := 0; i < int(f.SolidRating); i++ {
    output += getEmoji(":SolidStar:")
  }
  for i := 0; i < int(f.FadedRating); i++ {
    output += getEmoji(":FadedStar:")
  }
  // Todo: add emojis for 1-6 of each star to reduce text length
  // for i := 0; i < 6 - int(f.SolidRating + f.FadedRating); i++ {
  //   output += getEmoji(":NoStar:")
  // }

  return output
}

func getEmoji(emoji string) string {
  return fmt.Sprintf("<%s%s>", emoji, emojiIds[emoji])
}

type dayForecast struct {
  ForecastStartTimestamp int64
  ForecastPeriods []mswclient.ForecastResult
}