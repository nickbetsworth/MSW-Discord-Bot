## Overview

A discord bot which provides chat commands to pull surf forecasts from Magicseaweed into a group chat.

## API

### Searching for spots
Query
> https://magicseaweed.com/api/mdkey/search?match=CONTAINS&type=SPOT&query=\*SEARCH_CRITERIA\*

Response
```
[
  type: "SPOT",
  results: [
    {
      name: string,
      URL: string,
      id: int,
      score: int
    }
  ]
]
```

### Querying spot data
Query
> http://magicseaweed.com/api/mdkey/forecast/?spot_id=\*SPOT_ID\*&fields=solidRating,fadedRating,swell.*,timestamp

Response
```
[
  {
    timestamp: int,
    fadedRating: int,
    solidRating: int,
    swell: {
      height: int,
      absHeight: float,
      direction: float,
      trueDirection: float,
      period: int,
      unit: string,
      minBreakingHeight: int,
      maxBreakingHeight: int
    }
  }
]
```

## Typical use case

> /msw "spot name"

This chat command returns an ordered list of matching spot names. The user then selects a specific spot. If there is only a single match, the spot is automatically selected.

The user selects the second matched spot:
> ;s 2


## Todo
- Factor out units into command line argument
- Resolve time zone issues
- Allow a user to query for a particular day