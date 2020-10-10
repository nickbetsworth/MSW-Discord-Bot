package mswclient

import (
  "encoding/json"
  "io/ioutil"
  "net/http"
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

// GetSpots returns a list of spots matching the given query string
func GetSpots(query string) SpotResults {
  response, err := http.Get("https://magicseaweed.com/api/mdkey/search?match=CONTAINS&type=SPOT&query=" + query)

  if err != nil {
    // Todo: ensure we still catch this error further up the stack
    panic(err)
  }

  data, _ := ioutil.ReadAll(response.Body)

  var result searchResult

  json.Unmarshal([]byte(data), &result)

  return result[0].SpotResults
}