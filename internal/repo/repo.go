package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

const (
	LanguagesFile = "languages.yaml"
	PublishFile   = "publish.yaml"
)

var TopicOrder = []string{"geography", "space", "animals", "sports", "science"}

var Topics = map[string]struct{}{
	"geography": {},
	"space":     {},
	"animals":   {},
	"sports":    {},
	"science":   {},
}

type LanguagesConfig struct {
	Default   string                  `yaml:"default"`
	Languages map[string]LanguageMeta `yaml:"languages"`
	Order     []string                `yaml:"-"`
}

type LanguageMeta struct {
	Name string `yaml:"name"`
}

type PublishConfig struct {
	Defaults map[string]any         `yaml:"defaults"`
	Generate GenerateDefaults       `yaml:"generate"`
	Topics   map[string]TopicConfig `yaml:"topics"`
}

type GenerateDefaults struct {
	Defaults struct {
		Cards int `yaml:"cards"`
	} `yaml:"defaults"`
}

type TopicConfig struct {
	Enabled   *bool    `yaml:"enabled"`
	Rate      int      `yaml:"rate"`
	AgeBands  []string `yaml:"age_bands"`
	Subtopics []string `yaml:"subtopics"`
	Generate  struct {
		Enabled *bool `yaml:"enabled"`
		Cards   int   `yaml:"cards"`
	} `yaml:"generate"`
}

func (t TopicConfig) IsEnabled() bool {
	return t.Enabled == nil || *t.Enabled
}

func (t TopicConfig) GenerateEnabled() bool {
	return t.Generate.Enabled == nil || *t.Generate.Enabled
}

func FindRoot() (string, error) {
	if env := os.Getenv("CARTITAS_ROOT"); env != "" {
		return filepath.Abs(env)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if root, ok := walkForRoot(cwd); ok {
		return root, nil
	}
	exe, err := os.Executable()
	if err == nil {
		if root, ok := walkForRoot(filepath.Dir(exe)); ok {
			return root, nil
		}
	}
	return "", fmt.Errorf("could not find repo root (looked for %s from %s)", LanguagesFile, cwd)
}

func walkForRoot(start string) (string, bool) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, LanguagesFile)); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func LoadLanguages(root string) (*LanguagesConfig, error) {
	return LoadLanguagesFile(filepath.Join(root, LanguagesFile))
}

func LoadLanguagesFile(path string) (*LanguagesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg LanguagesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Default == "" {
		cfg.Default = "en"
	}
	cfg.Order = languageOrder(data, cfg.Languages)
	return &cfg, nil
}

func languageOrder(data []byte, langs map[string]LanguageMeta) []string {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil || len(root.Content) == 0 {
		return fallbackLangOrder(langs)
	}
	doc := root.Content[0]
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if doc.Content[i].Value != "languages" {
			continue
		}
		m := doc.Content[i+1]
		var order []string
		for j := 0; j+1 < len(m.Content); j += 2 {
			order = append(order, m.Content[j].Value)
		}
		if len(order) > 0 {
			return order
		}
	}
	return fallbackLangOrder(langs)
}

func fallbackLangOrder(langs map[string]LanguageMeta) []string {
	preferred := []string{"en", "es", "de"}
	seen := map[string]bool{}
	var ordered []string
	for _, p := range preferred {
		if _, ok := langs[p]; ok {
			ordered = append(ordered, p)
			seen[p] = true
		}
	}
	for c := range langs {
		if !seen[c] {
			ordered = append(ordered, c)
		}
	}
	return ordered
}

func LanguageCodes(cfg *LanguagesConfig) []string {
	if len(cfg.Order) > 0 {
		return append([]string{}, cfg.Order...)
	}
	return fallbackLangOrder(cfg.Languages)
}

func LoadPublish(root string) (*PublishConfig, error) {
	return LoadPublishFile(filepath.Join(root, PublishFile))
}

func LoadPublishFile(path string) (*PublishConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg PublishConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func DefaultsCardsPerPublish(cfg *PublishConfig) int {
	if cfg.Defaults == nil {
		return 2
	}
	v, ok := cfg.Defaults["cards_per_publish"]
	if !ok {
		return 2
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 2
	}
}

func TopicDirName(topicKey string) string {
	parts := strings.FieldsFunc(topicKey, func(r rune) bool {
		return r == '_' || unicode.IsSpace(r)
	})
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(strings.ToLower(p))
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	return b.String()
}
