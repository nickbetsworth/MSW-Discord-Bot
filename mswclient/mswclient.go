package mswclient

import (
  "encoding/json"
  "io/ioutil"
  "net/http"
  "fmt"
)

type SpotResults []struct {
  ID               int64   `json:"id"`
  URL              string  `json:"URL"`
  Name             string  `json:"name"`
  Score            float64 `json:"score"`
} 

type searchResult []struct {
  SpotResults SpotResults `json:"results"`
  Type string `json:"type"`
}

type ForecastResults []struct {
  FadedRating int64 `json:"fadedRating"`
  SolidRating int64 `json:"solidRating"`
  Swell       struct {
    Height    float64 `json:"height"`
    Period    int64   `json:"period"`
    Unit      string  `json:"unit"`
  } `json:"swell"`
  Timestamp int64 `json:"timestamp"`
}

// GetSpots returns a list of spots matching the given query string
func GetSpots(query string) SpotResults {
  response, err := http.Get("https://magicseaweed.com/api/mdkey/search?match=CONTAINS&type=SPOT&query=" + query)

  if err != nil {
    // Todo: ensure we still catch this error further up the stack
    panic(err)
  }

  data, _ := ioutil.ReadAll(response.Body)

  var results searchResult
  json.Unmarshal([]byte(data), &results)

  return results[0].SpotResults
}

func GetForecast(spotId int64) ForecastResults {
  var forecastEndpoint = fmt.Sprintf("http://magicseaweed.com/api/mdkey/forecast/?spot_id=%d&fields=solidRating,fadedRating,swell.*,timestamp", spotId)
  response, err := http.Get(forecastEndpoint)

  if err != nil {
    panic(err)
  }

  data, _ := ioutil.ReadAll(response.Body)

  var results ForecastResults
  json.Unmarshal([]byte(data), &results)

  return results
}