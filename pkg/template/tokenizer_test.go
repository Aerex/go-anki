package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testTokenInfo struct {
	text      string
	tokenType TokenType
	hasNext   bool
}

func TestTokenizer(t *testing.T) {
	tests := []struct {
		Name       string
		Template   string
		isFailTest bool
		TokenInfo  []testTokenInfo
	}{
		{
			Name:     "template with field replacemets",
			Template: "{{ Field_0 }}",
			TokenInfo: []testTokenInfo{
				{
					text:      "Field_0",
					tokenType: FieldReplacementToken,
					hasNext:   false,
				},
			},
		},
		{
			Name:     "template with open conditional",
			Template: "{{#Back}}",
			TokenInfo: []testTokenInfo{
				{
					text:      "Back",
					tokenType: OpenConditionalToken,
					hasNext:   false,
				},
			},
		},
		{
			Name:     "template with close conditional",
			Template: "{{/Back}}",
			TokenInfo: []testTokenInfo{
				{
					text:      "Back",
					tokenType: CloseConditionalToken,
					hasNext:   false,
				},
			},
		},
		{
			Name:     "template with plain text",
			Template: "Text",
			TokenInfo: []testTokenInfo{
				{
					text:      "Text",
					tokenType: TextToken,
					hasNext:   false,
				},
			},
		},
		{
			Name:     "template with negated conditional",
			Template: "{{^Text}}",
			TokenInfo: []testTokenInfo{
				{
					text:      "Text",
					tokenType: OpenNegatedToken,
					hasNext:   false,
				},
			},
		},
		{
			Name:     "template with conditional and text ",
			Template: "{{#Back}}Should Add Back Card{{/Back}}",
			TokenInfo: []testTokenInfo{
				{
					text:      "Back",
					tokenType: OpenConditionalToken,
					hasNext:   true,
				},
				{
					text:      "Should Add Back Card",
					tokenType: TextToken,
					hasNext:   true,
				},
				{
					text:      "Back",
					tokenType: CloseConditionalToken,
					hasNext:   false,
				},
			},
		},
	}

	for _, tt := range tests {
		tokenizer := NewTokenizer(tt.Template)
		for _, info := range tt.TokenInfo {
			tokenType := tokenizer.Next()
			if tokenType == EndOfBuffer {
				if !tt.isFailTest {
					assert.FailNowf(t, "", "Should not expect a failure for: %s", tt.Name)
				}
			}
			assert.Equal(t, info.text, tokenizer.Text(), tt.Name)
			assert.Equal(t, info.tokenType, tokenType, tt.Name)
			assert.Equal(t, info.hasNext, tokenizer.HasNext(), tt.Name)
		}
	}
}
