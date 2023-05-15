package template

import (
	"strings"

	"github.com/google/gapid/core/math/sint"
)

type TokenType uint32

const (
	OPENINGTAG = "{{"
	CLOSINGTAG = "}}"
)

const (
	EndOfBuffer TokenType = iota
	// {{ "Text" }}
	TextToken
	// {{ Field }}
	FieldReplacementToken
	// {{ #Field }}
	OpenConditionalToken
	// {{ /Field }}
	CloseConditionalToken
	// {{ ^Field }}
	OpenNegatedToken
)

// offset is the range of text for the contents within an anki template  `{{ }}`
type offset struct {
	start, end int
}

type Token struct {
	Type TokenType
	// text within {{ }}
	Content string
}

type Tokenizer struct {
	src   string
	cur   int
	inTag bool
	buf   offset
	tt    TokenType
	err   error
}

func NewTokenizer(src string) *Tokenizer {
	return &Tokenizer{
		src: src,
	}
}

func (tz *Tokenizer) Text() string {
	if tz.tt != EndOfBuffer {
		return tz.src[tz.buf.start:tz.buf.end]
	}
	return ""
}
func (tz *Tokenizer) HasNext() bool {
	return tz.cur < len(tz.src)
}

func (tz *Tokenizer) Next() TokenType {
	var skipWhiteSpaceCount int
	if tz.err != nil {
		tz.tt = EndOfBuffer
		return tz.tt
	}

	var char string
	for tz.cur < len(tz.src) {
		char = tz.src[tz.cur : tz.cur+1]
		if char == " " {
			// skip spaces
			tz.cur = tz.cur + 1
			if tz.inTag {
				skipWhiteSpaceCount = skipWhiteSpaceCount + 1
			}
			continue
		}
		if tz.src[tz.cur:tz.cur+2] == OPENINGTAG {
			// consume opening tag `` and skip to next char
			tz.inTag = true

			// if the current char is `` then only consume one char
			tz.cur = tz.cur + 2
			if char != "{" {
				tz.cur = tz.cur + 3
			}

			tz.buf.start = tz.cur
			continue
		}

		if tz.src[tz.cur:tz.cur+2] == CLOSINGTAG {
			tz.inTag = false
			tz.buf.end = tz.cur - skipWhiteSpaceCount

			tz.cur = tz.cur + 2
			if char != "}" {
				tz.cur = tz.cur + 3
			}

			return tz.tt
		}
		var idx int
		if tz.inTag {
			leadingChar := tz.src[tz.cur : tz.cur+1]
			switch leadingChar {
			case "#":
				tz.tt = OpenConditionalToken
			case "/":
				tz.tt = CloseConditionalToken
			case "^":
				tz.tt = OpenNegatedToken
			default:
				tz.tt = FieldReplacementToken
			}
			// read until the end of token
			idx = strings.Index(tz.src[tz.cur:], CLOSINGTAG)

			// buffer content starts after the leading character
			// only if token in not a field replacement
			if tz.tt != FieldReplacementToken {
				tz.buf.start = tz.cur + 1
			} else {
				tz.buf.start = tz.cur
			}

			// fast forward to end of token
			tz.cur = tz.cur + idx
		} else {
			tz.buf.start = tz.cur
			// Use the next opening tag to determine the end of buffer
			idx = strings.Index(tz.src[tz.cur:], OPENINGTAG)
			if idx < 0 {
				// if next open tag does not exist assume the end of buffer
				// is at the end of source string
				idx = len(tz.src)
				// Make sure we don't get index out of bounds by binding the end
				tz.buf.end = sint.Min(tz.cur+idx, idx)
			} else {
				tz.buf.end = tz.cur + idx
			}
			tz.tt = TextToken
			tz.cur = tz.cur + idx

			// If next position of the cursor exceeds the buffer
			// we should return EndOfBuffer
			if tz.buf.start > tz.buf.end {
				tz.tt = EndOfBuffer
			}
			return tz.tt
		}
	}

	tz.tt = EndOfBuffer
	return tz.tt

}
