package models

//import "time"

// Structure for schedule
type DeckSchedule struct {
  // The number of cards due for a given deck
  Due int `json:"due"`
  // The number of cards next for study
  Next int `json:"next"`
}
// Structure for generic data returned from API calls
type DataWithMeta struct {
  Data []struct {}
  Meta struct {}
}

// The strutucture representing sta
type Stats struct {
  StudiedToday `json:"studiedToday"`
}
// The structure representing the stats studied today
type StudiedToday struct {
  // The number of cards studied today
  Cards int `json:"cards"`
  // The number of seconds spent studying today
  Time int `json:"time"`
}
type CollectionStats struct {
  Stats `json:"stats"`
}

// Structure for deck
// See https://github.com/ankidroid/Anki-Android/wiki/Database-Structure#decks-jsonobjects
type Deck struct {
  // Deck unique ID (generated as long int)"
  //ID int8 `json:"id"`
  Name string `json:"name"`
  // Extended revision card limit (for custom study)
  // Potentially absent, in this case it's considered to be 10 by aqt.customstudy",
 // ExtendRev int `json:"extendedRev"`
 // // The unique sequence number
 // // TODO: What is a usn do?
 // Usn int `json:"usn"`
 // // True when deck is collapsed
  Collapsed bool `json:"collapsed"`
 // // True when deck collapsed in browser
 // BrowserCollapsed bool `json:"browserCollapsed"`
 // // The number of days that have passed between the collection was created and the deck was last updated from today
 // // First number is always 0
 // NewToday []int `json:"newToday"`
 // RevToday []int `json:"revToday"`
 // LrnToday []int `json:"lrnToday"`
 // // True if deck is dynamic (AKA filtered)
 // Dyn bool `json:"dyn"`
 // // Extended new card limit (for custom study).
 // ExtendNew int `json:"extendedNew"`
 // // Id of option group from the deck. 0 if the deck is dynamic
 // Conf int `json:"conf"`
 // // Last modification number
 // Mod time.Time `json:"mod"`
 // // Deck description
 // Desc string `json:"desc"`
  Schedule DeckSchedule `json:"schedule"`
}
