package main

import (
	"fmt"
  "github.com/nickbetsworth/mswclient"
)

func main() {
  var spots mswclient.SpotResults = mswclient.GetSpots("Porthcawl")
  fmt.Println(spots[0].ID)
}
