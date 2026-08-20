package webexport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/qu1queee/cartitas/internal/cardsplit"
	"github.com/qu1queee/cartitas/internal/repo"
)

var (
	headingLine = regexp.MustCompile(`(?m)^#.*\n?`)
	metaLine    = regexp.MustCompile(`(?m)^<!--.*?-->\n?`)
	clozeRe     = regexp.MustCompile(`\[([^\]]+)\]`)
)

type Catalog struct {
	GeneratedAt string `json:"generatedAt"`
	Cards       []Card `json:"cards"`
}

type Card struct {
	ID       string `json:"id"`
	Lang     string `json:"lang"`
	Topic    string `json:"topic"`
	Subtopic string `json:"subtopic"`
	Age      string `json:"age"`
	Kind     string `json:"kind"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func CardsJSON(root string) ([]byte, error) {
	cat, err := BuildCatalog(root)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(cat, "", "  ")
}

func BuildCatalog(root string) (Catalog, error) {
	cfg, err := repo.LoadLanguages(root)
	if err != nil {
		return Catalog{}, err
	}
	var cards []Card
	for _, lang := range repo.LanguageCodes(cfg) {
		base := filepath.Join(root, "cards", lang)
		err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return err
			}
			parsed, err := parseFile(root, lang, path)
			if err != nil {
				return err
			}
			cards = append(cards, parsed...)
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return Catalog{}, err
		}
	}
	sort.Slice(cards, func(i, j int) bool { return cards[i].ID < cards[j].ID })
	return Catalog{GeneratedAt: time.Now().UTC().Format(time.RFC3339), Cards: cards}, nil
}

func WriteCards(root string) (string, int, error) {
	data, err := CardsJSON(root)
	if err != nil {
		return "", 0, err
	}
	out := filepath.Join(root, "docs", "data", "cards.json")
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return "", 0, err
	}
	if err := os.WriteFile(out, append(data, '\n'), 0o644); err != nil {
		return "", 0, err
	}
	var cat Catalog
	_ = json.Unmarshal(data, &cat)
	rel, _ := filepath.Rel(root, out)
	return filepath.ToSlash(rel), len(cat.Cards), nil
}

func parseFile(root, lang, path string) ([]Card, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rel, _ := filepath.Rel(root, path)
	rel = filepath.ToSlash(rel)
	topic := filepath.Base(filepath.Dir(path))
	base := strings.TrimSuffix(filepath.Base(path), ".md")
	subtopic, age := splitName(base)

	body := string(data)
	body = headingLine.ReplaceAllString(body, "")
	body = metaLine.ReplaceAllString(body, "")
	var cards []Card
	for i, block := range cardsplit.Split(body) {
		parsed := parseBlock(block)
		for j, c := range parsed {
			c.Lang = lang
			c.Topic = topic
			c.Subtopic = subtopic
			c.Age = age
			c.ID = fmt.Sprintf("%s:%d:%d", rel, i+1, j+1)
			cards = append(cards, c)
		}
	}
	return cards, nil
}

func splitName(base string) (subtopic, age string) {
	i := strings.LastIndex(base, "_")
	if i <= 0 {
		return base, ""
	}
	return base[:i], base[i+1:]
}

func parseBlock(block string) []Card {
	lines := strings.Split(block, "\n")
	var q, a, c []string
	mode := ""
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "Q:"):
			mode = "q"
			q = append(q, strings.TrimSpace(strings.TrimPrefix(line, "Q:")))
		case strings.HasPrefix(line, "A:"):
			mode = "a"
			a = append(a, strings.TrimSpace(strings.TrimPrefix(line, "A:")))
		case strings.HasPrefix(line, "C:"):
			mode = "c"
			c = append(c, strings.TrimSpace(strings.TrimPrefix(line, "C:")))
		default:
			switch mode {
			case "q":
				q = append(q, line)
			case "a":
				a = append(a, line)
			case "c":
				c = append(c, line)
			}
		}
	}
	if len(c) > 0 {
		return parseCloze(strings.TrimSpace(strings.Join(c, "\n")))
	}
	if len(q) > 0 && len(a) > 0 {
		return []Card{{
			Kind:     "qa",
			Question: strings.TrimSpace(strings.Join(q, "\n")),
			Answer:   strings.TrimSpace(strings.Join(a, "\n")),
		}}
	}
	return nil
}

func parseCloze(text string) []Card {
	matches := clozeRe.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return nil
	}
	var cards []Card
	for _, m := range matches {
		answer := text[m[2]:m[3]]
		q := text[:m[0]] + "_____" + text[m[1]:]
		q = clozeRe.ReplaceAllString(q, "$1")
		cards = append(cards, Card{Kind: "cloze", Question: q, Answer: answer})
	}
	return cards
}
