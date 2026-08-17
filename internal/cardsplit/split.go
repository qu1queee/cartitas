package cardsplit

import (
	"regexp"
	"strings"
)

var cardSep = regexp.MustCompile(`(?m)^---\s*$`)

func Split(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	parts := cardSep.Split(text, -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func Join(cards []string) string {
	if len(cards) == 0 {
		return ""
	}
	return strings.Join(cards, "\n\n---\n\n") + "\n"
}
