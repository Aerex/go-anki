package models

import "github.com/aerex/anki-cli/api/types"

// structure for cards
type Card struct {
	Id       types.UnixTime `json:"id"`
	Nid      types.UnixTime `json:"nid"`
	Did      types.UnixTime `json:"did"`
	Ord      int            `json:"ord"`
	Mod      types.UnixTime
	Usn      int `json:"usn"`
	Type     int
	Queue    int
	Due      types.UnixTime
	Ivl      int `json:"lvl"`
	Factor   int `json:"factor"`
	Reps     int `json:"reps"`
	Lapses   int `json:"lapses"`
	Left     int `json:"left"`
	Odue     int
	Flags    int
	Question string `json:"question"`
	Answer   string `json:"answer"`

	IsEmpty bool `json:"isEmpty"`
	Name    string
	Note    struct {
		Fields []string
		Flags  int
		Guid   string
		Id     types.UnixTime
		Mid    types.UnixTime
		Mod    types.UnixTime
		// model describes the note type
		Model struct {
			Css  string
			Name string
			Did  types.UnixTime
			Flds []struct {
				Font   string
				Media  []string
				Name   string
				Ord    int
				Size   int
				Sticky bool
			} `json:"flds"`
		}
		LatexPost string
		LatexPre  string `json:"latexPost,omitempty"`
	} `json:"note"`
	Deck struct {
		Name string `json:"name"`
	} `json:"deck"`
}

// Structure for schedule
type DeckSchedule struct {
	// The number of cards due for a given deck
	Due int `json:"due"`
	// The number of cards next for study
	Next int `json:"next"`
}

// Structure for generic data returned from API calls
type DataWithMeta struct {
	Data []struct{}
	Meta struct{}
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
	ID   int    `json:"id"`
	Name string `json:"name"`
	// Extended revision card limit (for custom study)
	// Potentially absent, in this case it's considered to be 10 by aqt.customstudy",
	ExtendRev int `json:"extendedRev"`
	// The update sequence number used to figure out diffs when syncing.
	// value of -1 indicates changes that need to be pushed to server.
	// usn < server usn indicates changes that need to be pulled from server.
	Usn int `json:"usn"`
	// True when deck is collapsed
	Collapsed bool `json:"collapsed"`
	// True when deck collapsed in browser
	BrowserCollapsed bool `json:"browserCollapsed"`
	// The number of days that have passed between the collection was created and the deck was last updated from today
	// First number is always 0
	NewToday []int `json:"newToday"`
	RevToday []int `json:"revToday"`
	LrnToday []int `json:"lrnToday"`
	// True if deck is dynamic (AKA filtered)
	Dyn bool `json:"dyn"`
	// Extended new card limit (for custom study).
	ExtendNew int `json:"extendedNew"`
	// Id of option group from the deck. 0 if the deck is dynamic
	Conf int `json:"conf"`
	// Last modification number
	Mod types.UnixTime `json:"mod"`
	// Deck description
	Desc     string       `json:"desc"`
	Schedule DeckSchedule `json:"schedule"`
}

type LeechActionType int

const (
	LeechActionSuspend int = iota
	LeechActionMark
)

type OrderType int

const (
	NewCardsRandom OrderType = iota
	NewCardsDue
)

// structure for deck options
// See https://github.com/ankidroid/Anki-Android/wiki/Database-Structure#dconf-jsonobjects
type DeckConfig struct {
	// Whether the audio associated to a question should be
	Autoplay bool `json:"autoplay"`
	// Whether this deck is dynamic. Not present by default in decks.py
	Dyn bool `json:"dyn"`
	// The deck's ID
	DeckId int `json:"deckId"`
	// The configuration for lapse cards.
	Lapse struct {
		// The list of successive delay between the learning steps of the new cards
		Delays []int `json:"delays"`
		// What to do to leech cards.
		// Current values: 0 for suspend, 1 for mark.
		LeechAction LeechActionType `json:"leechAction"`
		// The number of lapses authorized before doing leechAction.
		LeechFails int `json:"leechFails"`
		// A lower limit to the new interval after a leech
		MinInterval int `json:"minInterval"`
		// Percent by which to multiply the current interval when a card goes has lapsed
		Mult int `json:"mult"`
	} `json:"lapse"`
	// The number of seconds after which to stop the timer
	MaxTaken int `json:"maxTaken"`
	// Last modification time
	Mod types.UnixTime `json:"mod"`
	// The name of the configuration
	Name string `json:"name"`
	// The configuration for new cards.
	New struct {
		// Whether to bury cards related to new cards answered
		Bury bool `json:"bury"`
		// The list of successive delay between the learning steps of the new cards
		Delays []int `json:"delays"`
		// The initial ease factor
		InitialFactor int `json:"initialFactor"`
		// The list of delays according to the button pressed while leaving the learning mode.
		// Good, easy and unused. In the GUI, the first two elements corresponds to Graduating Interval and Easy interval
		LearningDelays []int `json:"learningDelays"`
		// In which order new cards must be shown. NEW_CARDS_RANDOM = 0 and NEW_CARDS_DUE = 1.
		Order OrderType `json:"order"`
		// Maximal number of new cards shown per day.
		PerDay int `json:"perDay"`
	} `json:"new"`
	// Whether the audio associated to a question should be played when the answer is shown
	Replayq bool `json:"relayq"`
	// The configuration for review cards.
	Rev struct {
		// Whether to bury cards related to new cards answered
		Bury bool `json:"bury"`
		// The number to add to the easyness when the easy button is pressed
		Ease4 int `json:"ease4"`
		// The new interval is multiplied by a random number between -fuzz and fuzz
		Fuzz int `json:"fuzz"`
		// Multiplication factor applied to the intervals Anki generates"
		IvlFct int `json:"ivlFct"`
		// The maximal interval for review
		MaxIvl int `json:"maxIvl"`
		// Numbers of cards to review per day
		PerDay int `json:"perDay"`
	} `json:"rev"`
	// Whether timer should be shown
	Timer bool `json:"timer"`
	// TODO: I think this should be readonly
	// Update sequence number used to figure out diffs when syncing.
	// value of -1 indicates changes that need to be pushed to server.
	// usn < server usn indicates changes that need to be pulled from server.
	Usn int `json:"usn"`
}
