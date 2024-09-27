package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type UnixTime int64
type BoolVar bool
type TimestampSeconds time.Time

type ID int64

func (t *UnixTime) Scan(src interface{}) error {
	var ut int64
	switch x := src.(type) {
	case float64:
		ut = int64(src.(float64))
	case int64:
		ut = src.(int64)
	case nil:
		return nil
	default:
		return fmt.Errorf("Incompatible type for TimestampSeconds: %s", x)
	}
	*t = UnixTime(ut)
	return nil
}

func (t *UnixTime) UnmarshalJSON(src []byte) error {
	var ts interface{}
	if err := json.Unmarshal(src, &ts); err != nil {
		return err
	}
	return t.Scan(ts)
}

func (i *ID) UnmarshalJSON(b []byte) error {
	var id interface{}
	if err := json.Unmarshal(b, &id); err != nil {
		return err
	}
	return i.Scan(id)

}

// Scan implements the sql.Scanner interface for the ID type.
func (i *ID) Scan(src interface{}) error {
	var id int64
	switch x := src.(type) {
	case float64:
		id = int64(src.(float64))
	case int64:
		id = src.(int64)
	case string:
		var err error
		id, err = strconv.ParseInt(src.(string), 10, 64)
		if err != nil {
			return err
		}
	case nil:
		return nil
	default:
		return fmt.Errorf("Incompatible type for ID: %s", x)
	}
	*i = ID(id)
	return nil
}

func scanBlob(src interface{}, dst interface{}, tag string) error {
	var blob []byte
	switch src.(type) {
	case []byte:
		blob = src.([]byte)
	case string:
		blob = []byte(src.(string))
	default:
		return fmt.Errorf("Incompatible type for map struct %s", tag)
	}
	return json.Unmarshal(blob, dst)
}

func (nt *NoteTypes) Scan(src interface{}) error {
	return scanBlob(src, nt, "NoteTypes")
}

func (c *CollectionConf) Scan(src interface{}) error {
	return scanBlob(src, c, "CollectionConf")
}

func (dc *DeckConfigs) Scan(src interface{}) error {
	return scanBlob(src, dc, "DeckConf")
}

func (t *TagCache) Scan(src interface{}) error {
	return scanBlob(src, t, "TagCache")
}

// TODO: Abstract into on method for NoteTypes and Decks
func (nt *NoteTypes) UnmarshalJSON(src []byte) error {
	// unmarshall raw data using string as ID
	// then create a new map with ID as int64 (models.type.ID)
	t := make(map[string]*NoteType)
	if err := json.Unmarshal(src, &t); err != nil {
		return err
	}
	m := make(map[ID]*NoteType)
	for _, v := range t {
		m[v.ID] = v
	}
	*nt = NoteTypes(m)
	return nil
}

// Scan implements the sql.Scanner interface for the Decks type.
func (d *Decks) Scan(src interface{}) error {
	return scanBlob(src, d, "Deck")
}

func (d *Decks) UnmarshalJSON(src []byte) error {
	// unmarshall raw data using string as ID
	// then create a new map with ID as int64 (models.type.ID)
	t := make(map[string]*Deck)
	if err := json.Unmarshal(src, &t); err != nil {
		return err
	}
	m := make(map[ID]*Deck)
	for _, v := range t {
		m[v.ID] = v
	}
	*d = Decks(m)
	return nil
}

func (d *DeckConfigs) UnmarshalJSON(src []byte) error {
	// unmarshall raw data using string as ID
	// then create a new map with ID as int64 (models.type.ID)
	t := make(map[string]*DeckConfig)
	if err := json.Unmarshal(src, &t); err != nil {
		return err
	}
	m := make(map[ID]*DeckConfig)
	for _, v := range t {
		m[v.ID] = v
	}
	*d = DeckConfigs(m)
	return nil
}

// Scan implements the sql.Scanner interface for the BoolVar type.
func (b *BoolVar) Scan(src interface{}) error {
	var tf bool
	switch t := src.(type) {
	case bool:
		tf = t
	case float64:
		// Only 0 is false
		tf = t != 0
	case int64:
		// Only 0 is false
		tf = t != 0
	case nil:
		// Nil is false
		tf = false
	default:
		return fmt.Errorf("Incompatible type '%T' for BoolVar", src)
	}
	*b = BoolVar(tf)
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for the BoolVar
// type.
func (b *BoolVar) UnmarshalJSON(src []byte) error {
	var tmp interface{}
	if err := json.Unmarshal(src, &tmp); err != nil {
		return err
	}
	return b.Scan(tmp)
}
func (f *NoteFields) Scan(src interface{}) error {
	var tmp string
	switch src.(type) {
	case []byte:
		tmp = string(src.([]byte))
	case string:
		tmp = src.(string)
	default:
		return errors.New("Incompatible type for NoteFields")
	}

	tmp = strings.ReplaceAll(tmp, "\n", "")
	// the values of the fields in this note. separated by 0x1f (31) character `^_`
	// FIXME: sometimes the value will be in unicode and sometimes it isn't.
	if strings.Contains(tmp, "^_") {
		*f = NoteFields(strings.Split(tmp, "^_"))
	} else {
		*f = NoteFields(strings.Split(tmp, "\x1f"))
	}
	return nil
}

func (f NoteFields) Value() (driver.Value, error) {
	return strings.Join(f, "\x1f"), nil
}

func (dc *DeckConfig) Get(key string, dflt interface{}) (reflect.Value, error) {
	s := reflect.ValueOf(dc)
	val := s.FieldByName(key)
	if !val.IsValid() {
		cval := reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem()
		if !cval.IsValid() {
			return reflect.Value{}, fmt.Errorf("could not determine value from %T", dflt)
		}
		return cval, nil
	}
	return val, nil
}
