package queries

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aerex/go-anki/api/sql/sqlite/repositories"
)

type Builder interface {
	Query() (cls string, args []string, err error)
}

type builder struct {
	qs       string
	tokens   []string
	colRepo  repositories.ColRepo
	deckRepo repositories.DeckRepo
	noteRepo repositories.NoteRepo
}

type queryState struct {
	isNot  bool
	isOr   bool
	isJoin bool
	cls    bytes.Buffer
	bad    bool
}

func NewBuilder(qs string, c repositories.ColRepo, d repositories.DeckRepo, n repositories.NoteRepo) Builder {
	return &builder{
		qs:       qs,
		colRepo:  c,
		deckRepo: d,
		noteRepo: n,
	}
}

func (s *queryState) add(query string, wrap bool) {
	if query == "" {
		if s.isNot {
			s.isNot = false
		} else {
			s.bad = true
		}
		return
	}

	if strings.EqualFold("skip", query) {
		return
	}

	if s.isJoin {
		if s.isOr {
			s.cls.WriteString(" or ")
			s.isJoin = false
		} else {
			s.cls.WriteString(" and ")
		}

		s.cls.WriteString(query)
		s.isJoin = true
	}

	if s.isNot {
		s.cls.WriteString(" not ")
		s.isNot = false
	}

	if wrap {
		s.cls.WriteString(fmt.Sprintf("(%s)", query))
	}
}

// TODO: Figure out how to handle errors
// Query generates a SQL clause with optional arguments using
// Anki query system. See https://docs.ankiweb.net/searching.html for more information
func (b *builder) Query() (cls string, args []string, err error) {
	b.tokenize()
	state := queryState{}
	sqlCls := clause{
		colRepo:  b.colRepo,
		deckRepo: b.deckRepo,
		noteRepo: b.noteRepo,
	}
	for _, token := range b.tokens {
		if state.bad {
			return "", []string{}, fmt.Errorf("invalid query: %s", strings.Join(b.tokens, " "))
		}
		// special tokens
		if token == "-" {
			state.isNot = true
		} else if strings.ToLower(token) == "or" {
			state.isOr = true
		} else if token == "(" {
			state.add(token, false)
			state.isJoin = false
		} else if token == ")" {
			state.cls.WriteString(")")
			// commands
		} else if strings.Contains(token, ":") {
			parts := strings.Split(token, ":")
			if len(parts) < 2 {
				return "", []string{}, fmt.Errorf("expected clause to be in format 'cmd:value' but received %s", token)
			}
			cmd := strings.ToLower(parts[0])
			val := parts[1]
			sqlCls.val = val
			sqlCls.cmd = cmd
			switch cmd {
			case "added":
				state.add(sqlCls.added(), true)
				break
			case "card":
				state.add(sqlCls.template(), true)
				break
			case "deck":
				state.add(sqlCls.deck(), true)
				break
			case "flag":
				state.add(sqlCls.flag(), true)
				break
			case "mid":
				state.add(sqlCls.mid(), true)
				break
			case "nid":
				state.add(sqlCls.nid(), true)
				break
			case "cid":
				state.add(sqlCls.cid(), true)
				break
			case "note":
				state.add(sqlCls.note(), true)
				break
			case "prop":
				state.add(sqlCls.prop(), true)
				break
			case "rated":
				state.add(sqlCls.rated(), true)
				break
			case "tag":
				state.add(sqlCls.tag(), true)
			case "dupe":
				state.add(sqlCls.dupes(), true)
				break
			case "is":
				state.add(sqlCls.cardState(), true)
				break
			default:
				state.add(sqlCls.field(), true)
				break
			}
		} else {
			// normal text search
			state.add(sqlCls.text(token), true)
		}
	}

	if state.bad {
		return "", []string{}, fmt.Errorf("invalid query: %s", strings.Join(b.tokens, " "))
	}
	return state.cls.String(), sqlCls.args, nil
}

func (b *builder) tokenize() {
	var inQuote rune
	var token bytes.Buffer
	var tokens []string
	for _, c := range b.qs {
		if c == '\'' || c == '"' {
			if inQuote != 0 {
				if c == inQuote {
					inQuote = 0
				} else {
					token.WriteRune(c)
				}
			} else if token.Len() != 0 {
				if strings.HasSuffix(token.String(), ":") {
					inQuote = c
				} else {
					token.WriteRune(c)
				}
			} else {
				inQuote = c
			}
			// separator
		} else if c == ' ' {
			if inQuote != 0 {
				token.WriteRune(c)
			} else if token.Len() != 0 {
				// space marks token finished
				tokens = append(tokens, token.String())
				token.Reset()
			}
			// nesting
		} else if c == '(' || c == ')' {
			if inQuote != 0 {
				token.WriteRune(c)
			} else {
				if c == ')' && token.Len() != 0 {
					tokens = append(tokens, token.String())
					token.Reset()
				}
				tokens = append(tokens, string(c))
			}
			// negation
		} else if c == '-' {
			if token.Len() != 0 {
				token.WriteRune(c)
			} else if len(tokens) == 0 || !strings.EqualFold(tokens[len(tokens)-1], "-") {
				tokens = append(tokens, "-")
			}
			// normal character
		} else {
			token.WriteRune(c)
		}
	}
	if token.Len() != 0 {
		tokens = append(tokens, token.String())
	}
	b.tokens = tokens
}
