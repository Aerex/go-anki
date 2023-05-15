package models

type CardQA struct {
	Question string
	Answer   string
	CardType string
	Deck     string
	Due      UnixTime
}

// structure for the card fields
type CardField struct {
	// Name of the field
	Name string `json:"name" yaml:"name"`
	// Sticky fields retain the value that was last added when adding new notes
	Sticky bool `json:"sticky" yaml:"stick"`
	// Determine if the field uses Right-to-Left
	RTL bool `json:"rtl" yaml:"rtl"`
	/* Identifies which of the card templates or cloze deletions it corresponds to
	   for card templates, valid values are from 0 to num templates - 1
	   for cloze deletions, valid values are from 0 to max cloze index - 1
	*/
	Ordinal int `json:"ord" yaml:"ord"`
	// Display font
	Font string `json:"font" yaml:"font"`
	// Font size
	FontSize int `json:"size" yaml:"size"`
}

type CardTemplate struct {
	// Template name
	Name string `json:"name"`
	// Same number ref in CardModel.Ordinal
	Ordinal int `json:"ord"`
	// Question format
	QuestionFormat string `json:"qfmt"`
	// Answer format
	AnswerFormat string `json:"afmt"`
	// Browser question format
	BrowserQuestionFormat string `json:"bqfmt"`
	// Browser answer format
	BrowserAnswerFormat string `json:"bafmt"`
	// Deck override for template represented by the deck id. Defaults to null
	DeckOverride ID `json:"did,omitempty"`
}

// Structure to determine if a card is generated or not and
// in what order the card is generated
// Used in modern clients. May exist for backwards compatibility.
// @see https://forums.ankiweb.net/t/is-req-still-used-or-present/9977 for more information
type CardRequirements struct {
	Ordinal int
	// none - no cards are generated for this template. The list should be empty
	// all - the card is generated only if each field of the list are filled
	// any - the card is generated if any of the field of the list is filled.
	CardGenerationType string
	Fields             []int
}

// Unmarshal card requirements as generic arrays
// Credits to https://github.com/flimzy/anki/blob/master/anki_types.go#L148
//func (req *CardRequirements) UnmarshalJSON(b []byte) error {
//	tmp := make([]interface{}, 3)
//	if err := json.Unmarshal(b, &tmp); err != nil {
//		return err
//	}
//	req.Ordinal = int(tmp[0].(float64))
//	req.CardGenerationType = tmp[1].(string)
//	tmpAry := tmp[2].([]interface{})
//	req.Fields = make([]int, len(tmpAry))
//	for i, v := range tmpAry {
//		req.Fields[i] = int(v.(float64))
//	}
//	return nil
//	// create a general interface of size 3 to
//	// mimic a generic array struc that which will have 3 elements
//	//	anonInterface := make([]interface{}, 3)
//	//	if err := json.Unmarshal(b, &anonInterface); err != nil {
//	//		return err
//	//	}
//	//
//	//	req.Ordinal = int(anonInterface[0].(int))
//	//	req.CardGenerationType = string(anonInterface[1].(string))
//	//
//	//	// Get the array as a list of intefaces
//	//	arryOfFields := anonInterface[2].([]interface{})
//	//	// Create an empty array to fill with correct valuues
//	//	req.Fields = make([]int, len(arryOfFields))
//	//
//	//	for i := 0; i < len(arryOfFields); i++ {
//	//		req.Fields[i] = int(arryOfFields[i].(int))
//	//	}
//	//	return nil
//}

type NoteTypes map[ID]*NoteType

// structure for the model
// provides information on the card's structure such as the css and the fields
type NoteType struct {
	// The model id (timestamp)
	ID ID `yaml:"id" json:"id,omitempty"`
	// The name of the model (ie: Basic)
	Name string `yaml:"name" json:"name" `
	// The list of tags on the card
	Tags []string `json:"tags" yaml:"tags"`
	// The deck id attached to the card
	// DeckId types.UnixTime `json:"did"`
	// A list of fields on the card
	Fields []*CardField `json:"flds" yaml:"flds"`
	// Integer specifying which field is used for sorting in the browser
	SortField int `json:"sortf"`
	// A list of formatting for the fields on the card
	Templates []*CardTemplate `json:"tmpls"`
	// Integer specifying what type of model. 0 for standard, 1 for cloze
	Type ModelType `json:"type"`
	// Preamble for LaTeX expressions
	LatexPre string `json:"latexPre"`
	// String added to end of LaTeX expressions (usually \\end{document})
	LatexPost string `json:"latexPost"`
	// CSS, shared for all templates
	CSS string `json:"css"`
	// Modification time in seconds
	// TODO: Currently returns in seconds. Change it to miliseconds for seconds in api for consistency
	Mod UnixTime `json:"mod"`
	// Array of card requirements describing which fields are required and what fields should be generated for the card
	//FIXME: Figure out why unmarshalling nested array is failing
	// Issue it the decoded byte array is not providing the full array to unmarshal
	//RequiredFields *CardRequirements `json:"req"`
	// Same as the other usn
	USN int `json:"usn"`
}

type NoteFields []string

// structure for card notes
type Note struct {
	// Array of key value pairs of the field name and field value
	Fields  NoteFields `json:"fields" yaml:"fields" db:"flds"`
	Flags   int        `json:"flags" yaml:"flags"`
	GUID    string     `json:"guid" yaml:"guid" db:"guid"`
	ID      ID         `json:"id" yaml:"id" db:"id"`
	ModelID ID         `json:"mid" yaml:"mid" db:"mid"`
	Mod     UnixTime   `json:"mod" yaml:"mod" db:"mod"`
	// model describes the note type
	Model     NoteType `json:"model" yaml:"model"`
	LatexPost string
	// sort field: used for quick sorting and duplicate check.
	// The sort field is an integer so that when users are sorting on a field that contains only numbers,
	// they are sorted in numeric instead of lexical order. Text is stored in this integer field.
	SortField  string `json:"sfld" yaml:"sfld" db:"sfld"`
	USN        int    `json:"usn" yaml:"usn" db:"usn"`
	LatexPre   string `json:"latexPost,omitempty" yaml:"laxtexPost,omitempty" db:"laxtexPost,omitempty"`
	StringTags string `json:"string_tags" yaml:"string_tags" db:"tags"`
	Checksum   uint64 `yaml:"csum" db:"csum"`
}

// structure for cards
type Card struct {
	ID             ID       `json:"id" db:"id"`
	Ord            int      `json:"ord" db:"ord"`
	Mod            UnixTime `json:"mod" db:"mod"`
	USN            int      `json:"usn" db:"usn"`
	Type           CardType `json:"type" db:"type" `
	Queue          CardQue  `json:"queue" db:"queue"`
	Due            UnixTime `json:"due" db:"due"`
	Interval       int      `json:"ivl" db:"ivl"`
	Factor         int      `json:"factor" db:"factor"`
	Reps           int      `json:"reps" db:"reps"`
	Lapses         int      `json:"lapses" db:"lapses"`
	Left           int      `json:"left" db:"left"`
	Odue           int      `json:"odue" db:"odue"`
	Flags          string   `json:"-"`
	Data           string   `json:"-"`
	Question       string   `json:"question" db:"question"`
	Answer         string   `json:"answer" db:"answer"`
	IsEmpty        bool     `json:"isEmpty" db:"isEmpty"`
	Note           Note     `json:"note" db:"note"`
	NoteID         ID       `json:"nid" db:"nid"`
	OriginalDeckID ID       `json:"odid" db:"odid"`
	DeckID         ID       `json:"did" db:"did"`
	Deck           Deck     `json:"deck" db:"deck"`
}

type CreateNote struct {
	Type   string   `yaml:"type"`
	Deck   string   `yaml:"deck"`
	Fields []string `yaml:"fields"`
	Tags   []string `yaml:"tags"`
}

type CardType int

const (
	CardTypeNew CardType = iota
	CardTypeLearning
	CardTypeReview
)

type CardQue int

const (
	CardQueueSBuried    CardQue = -3
	CardQueueBuried     CardQue = -2
	CardQueueSuspended  CardQue = -1
	CardQueueNew        CardQue = 0
	CardQueueLearning   CardQue = 1
	CardQueueReview     CardQue = 2
	CardQueueRelearning CardQue = 3
)

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
	Cards int `json:"cards" db:"cards"`
	// The number of seconds spent studying today
	Time int64 `json:"time" db:"time"`
	// The number of failed cards
	Failed int `json:"failed" db:"failed"`
	// The number of learning cards
	Learning int `json:"learning" db:"learning"`
	// The number of reviewed cards
	Review int `json:"review" db:"review"`
	// The number of relearned cards
	Relearn int `json:"relearned" db:"relearned"`
	// The number of filtered cards
	Filter int `json:"filter" filter:"filter"`
}

// The structure representing the collection for the user
type Collection struct {
	// arbitrary number since there is only one row
	ID ID `json:"id"`
	// update sequence number: used for finding diffs when syncing.
	USN     int      `json:"usn" db:"usn"`
	Created UnixTime `json:"crt" db:"crt"`
	// json object containing configuration options that are synced.
	Conf      CollectionConf `json:"conf" db:"conf"`
	NoteTypes NoteTypes      `json:"models" db:"models"`
	DeckConfs DeckConfigs    `json:"dconf" db:"dconf"`
	Decks     Decks          `json:"decks" db:"decks"`
}

type NewCardOrder int

const (
	NewCardDue NewCardOrder = iota
	NewCardRandom
)

//type DeckConf struct {
//	Id  int64 `json:"id"`
//	New struct {
//		// In which order new cards must be shown.
//		// NEW_CARDS_RANDOM = 0 and NEW_CARDS_DUE = 1.
//		// @see NewCardOrder enum
//		Order NewCardOrder
//	}
//}

// A map of deck configurations with the id as the creation timestamp
// If it is the default dec the value will be 1
type DeckConfigs map[ID]*DeckConfig

type CollectionStats struct {
	Stats `json:"stats"`
}

type ReviewType int

const (
	ReviewTypeLearn ReviewType = iota
	ReviewTypeReview
	ReviewTypeRelearn
	ReviewTypeCram
)

type ReviewLog struct {
	ID           ID    `json:"id" db:"id"`
	CID          ID    `json:"cid" db:"cid"`
	USN          int   `json:"usn" db:"usn"`
	Ease         int   `json:"ease" db:"ease"`
	Interval     int   `json:"ivl" db:"ivl"`
	LastInterval int   `json:"lastIvl" db:"lastIvl"`
	Factor       int   `json:"factor" db:"factor"`
	Time         int64 `json:"time" db:"time"`
	Type         int
}

type CollectionConf struct {
	// This is the highest value of a due value of a new card.
	// It allows to decide the due number to give to the next note created.
	// (This is useful to ensure that cards are seen in order in which they are added.,
	NextPos     int `json:"nextPos"`
	CurrentDeck ID  `json:"curDeck"`
}

// Structure for deck
// See https://github.com/ankidroid/Anki-Android/wiki/Database-Structure#decks-jsonobjects
type Deck struct {
	// Deck unique Id (generated as long int)"
	ID   ID     `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	// Extended revision card limit (for custom study)
	// Potentially absent, in this case it's considered to be 10 by aqt.customstudy",
	ExtendReviewCardLimit int `json:"extendedRev" db:"extendedRev"`
	// Extended new card limit (for custom study).
	ExtendNewCardLimit int `json:"extendedNew" db:"extendedNew"`
	// The update sequence number used to figure out diffs when syncing.
	// value of -1 indicates changes that need to be pushed to server.
	// usn < server usn indicates changes that need to be pulled from server.
	USN int `json:"usn" db:"usn"`
	// True when deck is collapsed
	Collapsed bool `json:"collapsed" db:"collapsed"`
	// True when deck collapsed in browser
	BrowserCollapsed bool `json:"browserCollapsed"`
	// The number of days that have passed between the collection was created and the deck was last updated from today
	// First number is always 0
	NewToday    [2]int `json:"newToday" db:"newToday"`
	RevewsToday [2]int `json:"revToday" db:"revToday"`
	LearnToday  [2]int `json:"lrnToday" db:"lrnToday"`
	// True if deck is dynamic (AKA filtered)
	// TODO: Create a type that interprets 0/1 as bool
	Dyn BoolVar `json:"dyn" db:"dyn"`
	// Id of option group from the deck. 0 if the deck is dynamic
	Conf int `json:"conf" db:"conf"`
	// Last modification number
	Mod *UnixTime `json:"mod" db:"mod"`
	// Deck description
	Desc     string       `json:"desc" db:"desc"`
	Schedule DeckSchedule `json:"schedule"`
}

type Decks map[ID]*Deck

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

type ModelType int

const (
	StandardCardType ModelType = iota
	ClozeCardType
)

// structure for deck options
// See https://github.com/ankidroid/Anki-Android/wiki/Database-Structure#dconf-jsonobjects
type DeckConfig struct {
	ID ID `json:"id"`
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
	Mod UnixTime `json:"mod"`
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
		Ease4 float32 `json:"ease4"`
		// The new interval is multiplied by a random number between -fuzz and fuzz
		Fuzz float32 `json:"fuzz"`
		// Multiplication factor applied to the intervals Anki generates"
		IvlFct int `json:"ivlFct"`
		// The maximal interval for review
		MaxIvl int `json:"maxIvl"`
		// Numbers of cards to review per day
		PerDay int `json:"perDay"`
	} `json:"rev"`
	// Whether timer should be shown
	Timer BoolVar `json:"timer"`
	// TODO: I think this should be readonly
	// Update sequence number used to figure out diffs when syncing.
	// value of -1 indicates changes that need to be pushed to server.
	// usn < server usn indicates changes that need to be pulled from server.
	USN int `json:"usn"`
}
