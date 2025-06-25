package japanesetokenizing

import (
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

type Tokenizer struct {
	tokenizer *tokenizer.Tokenizer
}

// Initialize the tokenizer once
func Initialize() (*Tokenizer, error) {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, err
	}
	return &Tokenizer{tokenizer: t}, nil
}

// TokenizeJapanese returns a list of surface words (no DUMMY tokens)
func (t *Tokenizer) TokenizeJapanese(text string) []string {
	tokens := t.tokenizer.Tokenize(text)
	words := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token.Class == tokenizer.DUMMY {
			continue
		}
		words = append(words, token.Surface)
	}
	return words
}
