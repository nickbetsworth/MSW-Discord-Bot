package mswclient

import (
  "encoding/json"
  "io/ioutil"
  "net/http"
  "net/url"
  "fmt"
)

const mswAPIEndpoint string = "http://magicseaweed.com/api/mdkey"

// GetSpots returns a list of spots matching the given query string
func GetSpots(query string) SpotResults {
  var endpoint = getAPIURL(fmt.Sprintf("search?match=CONTAINS&fields=*,tideURL&type=SPOT&query=%s", url.QueryEscape(query)))
  response, err := http.Get(endpoint)

  if err != nil {
    panic(err)
  }

  data, _ := ioutil.ReadAll(response.Body)

  var results searchResult
  json.Unmarshal([]byte(data), &results)

  return results[0].SpotResults
}

func GetTides(spotId int64, start int64, end int64) TideResults {
  var endpoint = getAPIURL(fmt.Sprintf("tide/?spot_id=%d&start=%d&end=%d", spotId, start, end))

  response, err := http.Get(endpoint)

  if err != nil {
    panic(err)
  }

  data, _ := ioutil.ReadAll(response.Body)

  var results TideResults
  json.Unmarshal([]byte(data), &results)

  return results;
}

func GetForecast(spotId int64) ForecastResults {
  var endpoint = getAPIURL(fmt.Sprintf("forecast/?spot_id=%d&fields=solidRating,fadedRating,swell.*,wind.*,timestamp,localTimestamp,threeHourTimeText&units=uk", spotId))
  response, err := http.Get(endpoint)

  if err != nil {
    panic(err)
  }

  data, _ := ioutil.ReadAll(response.Body)

  var results ForecastResults
  json.Unmarshal([]byte(data), &results)

  return results
}

func getAPIURL(endpoint string) string {
  return fmt.Sprintf("%s/%s", mswAPIEndpoint, endpoint)
}

type SpotResult struct {
  ID               int64   `json:"id"`
  URL              string  `json:"URL"`
  TideURL          string  `json:"tideURL"`
  Name             string  `json:"name"`
  Score            float64 `json:"score"`
}

type SpotResults []struct {
  SpotResult
} 

type searchResult []struct {
  SpotResults SpotResults `json:"results"`
  Type string `json:"type"`
}

type TideResult struct {
  Tide    []struct {
    Shift          float64 `json:"shift"`
    State          string  `json:"state"`
    Timestamp      int64   `json:"timestamp"`
    TimezoneOffset int64   `json:"timezoneOffset"`
    Unixtime       int64   `json:"unixtime"`
  } `json:"tide"`
  Timestamp int64  `json:"timestamp"`
  Unit      string `json:"unit"`
}

type TideResults []struct {
  TideResult
}

type ForecastResult struct {
  FadedRating int64 `json:"fadedRating"`
  SolidRating int64 `json:"solidRating"`
  Swell       struct {
    MinBreakingHeight int64   `json:"minBreakingHeight"`
    MaxBreakingHeight int64   `json:"maxBreakingHeight"`
    Height            float64 `json:"height"`
    Period            int64   `json:"period"`
    Unit              string  `json:"unit"`
  } `json:"swell"`
  Wind        struct {
    Speed             int64   `json:"speed"`
    Unit              string  `json:"unit"`
    StringDirection   string  `json:"stringDirection"`
  }
  ThreeHourTimeText string `json:"threeHourTimeText"`
  LocalTimestamp int64 `json:"localTimestamp"`
  Timestamp int64 `json:"timestamp"`
}

type ForecastResults []struct {
  ForecastResult
}
