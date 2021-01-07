package main

import (
	"strings"
  "fmt"
  "os"
  "log"
  "os/signal"
  "syscall"
  "time"
  "github.com/bwmarrin/discordgo"
  "github.com/joho/godotenv"
  "github.com/nickbetsworth/mswclient"
)

func init() {
  err := godotenv.Load()
  if err != nil {
    log.Println("Unable to load .env file")
  }

  token = os.Getenv("DISCORD_BOT_KEY")
}

var token string

var emojiIds = map[string]string{
  ":S0F0N2:": "767332050079449108",
  ":S0F1N1:": "767332050126110720",
  ":S0F2N0:": "767332049866063883",
  ":S1F0N1:": "767332050079580200",
  ":S1F1N0:": "767332050125455420",
  ":S2F0N0:": "767332049816125474",
}

func main() {
  if token == "" {
    log.Fatal("No token was provided. Please set the token in your .env file")
    return
  }

  discord, err := discordgo.New("Bot " + token)

  if err != nil {
    log.Fatal("Error initialising discord session: ", err)
  }
  
  discordgo.MakeIntent(discordgo.IntentsGuilds | discordgo.IntentsGuildMessages)

  discord.AddHandler(messageCreate)

  err = discord.Open()

  if err != nil {
    log.Fatal("Error opening discord session: ", err, token)
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

    if len(searchString) == 0 {
      s.ChannelMessageSend(event.ChannelID, "Please enter a search string i.e `!msw spot name`")
      return
    }

    s.ChannelMessageSend(event.ChannelID, fmt.Sprintf("Searching for surf spots at \"%s\"", searchString))
    
    // Todo: add support for direct querying by spot ID

    spots := mswclient.GetSpots(searchString)
    
    if len(spots) == 0 {
      s.ChannelMessageSend(event.ChannelID, fmt.Sprintf("No spots found matching \"%s\"", searchString))
      return
    }

    spot := spots[0]
    s.ChannelMessageSend(event.ChannelID, "Showing forecast for " + spot.Name)
    forecast := mswclient.GetForecast(spot.ID)

    groupedForecasts := groupForecastsByDay(forecast)

    for _, dayForecast := range groupedForecasts {
      tides := mswclient.GetTides(spot.ID, dayForecast.ForecastStartTimestamp, dayForecast.ForecastStartTimestamp)

      tideMessage := convertTideToMessage(tides)
      forecastMessage := convertDayForecastToMessage(dayForecast)

      startOfDay := time.Unix(dayForecast.ForecastStartLocalTimestamp, 0)

      _, err := s.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
        Title: startOfDay.Format("Monday 02/01"),
        Fields: []*discordgo.MessageEmbedField{&tideMessage, &forecastMessage},
      })

      if (err != nil) {
        log.Println(err)
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

func convertTideToMessage(tides mswclient.TideResults) discordgo.MessageEmbedField {
  var result discordgo.MessageEmbedField
  result.Name = "Tides"

  if len(tides) > 1 || len(tides) == 0 {
    log.Printf("Unexpected number of tide results (%d)\n", len(tides))
    result.Value = "Unable to retrieve tide data"
    return result
  }

  var tideStrings []string

  for _, t := range tides[0].TideResult.Tide {
    tm := time.Unix(t.Unixtime, 0)
    tideStrings = append(tideStrings, fmt.Sprintf("%s %s", t.State, tm.Format("3:04pm")))
  }

  result.Value = strings.Join(tideStrings, " | ")

  return result
}

func convertForecastPeriodToString(f mswclient.ForecastResult) string {
  return fmt.Sprintf(
    "%s %d-%d%s %s %.1f%s %ds | %d%s %s\n",
    f.ThreeHourTimeText,
    f.Swell.MinBreakingHeight, 
    f.Swell.MaxBreakingHeight, 
    f.Swell.Unit, 
    getStarRatingString(int(f.SolidRating), int(f.FadedRating), int(6 - f.SolidRating - f.FadedRating)), 
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
  
  for _, forecastPeriod := range ungroupedForecasts {
    if forecastPeriod.ThreeHourTimeText == "12am" || len(groupedForecasts) == 0 {
      groupedForecasts = append(groupedForecasts, dayForecast{ForecastStartTimestamp: forecastPeriod.Timestamp, ForecastStartLocalTimestamp: forecastPeriod.LocalTimestamp})
    }

    currentDayForecast := &groupedForecasts[len(groupedForecasts)-1]
    currentDayForecast.ForecastPeriods = append(currentDayForecast.ForecastPeriods, forecastPeriod.ForecastResult)
  }

  return groupedForecasts
}

func getStarRatingString(solidRating int, fadedRating int, noRating int) string {
  if solidRating + fadedRating + noRating == 0 {
    return ""
  }

  numSolid, numFaded, numBlank := 0, 0, 0
  if solidRating > 0 {
    numSolid = min(2, solidRating)
  }
  if fadedRating > 0 {
    numFaded = min((2-numSolid), fadedRating)
  }
  if noRating > 0 {
    numBlank = min((2-numSolid-numFaded), noRating)
  }

  return getEmoji(fmt.Sprintf(":S%dF%dN%d:", numSolid, numFaded, numBlank)) + getStarRatingString(solidRating-numSolid, fadedRating-numFaded, noRating-numBlank)
}

func getEmoji(emoji string) string {
  return fmt.Sprintf("<%s%s>", emoji, emojiIds[emoji])
}

func min(a int, b int) int {
  if a < b {
    return a
  }

  return b
}

type dayForecast struct {
  ForecastStartTimestamp int64
  ForecastStartLocalTimestamp int64
  ForecastPeriods []mswclient.ForecastResult
}
