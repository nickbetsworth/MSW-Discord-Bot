package main

import (
	"fmt"
  "github.com/nickbetsworth/mswclient"
)

func main() {
  spots := mswclient.GetSpots("Porthcawl")
  fmt.Println(spots[0].ID)

  forecast := mswclient.GetForecast(spots[0].ID)
  fmt.Printf("Solid rating: %d\n", forecast[0].SolidRating)
  fmt.Printf("Faded rating: %d\n", forecast[0].FadedRating)
}
