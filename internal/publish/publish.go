package publish

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/qu1queee/cartitas/internal/cardsplit"
	"github.com/qu1queee/cartitas/internal/repo"
)

func Topic(root, topicKey string, topicCfg repo.TopicConfig, defaultsPerPublish int, lang string, dryRun bool) []string {
	rate := topicCfg.Rate
	if rate == 0 {
		rate = defaultsPerPublish
	}
	topicName := repo.TopicDirName(topicKey)
	queueDir := filepath.Join(root, "queue", lang, topicName)
	cardsDir := filepath.Join(root, "cards", lang, topicName)

	info, err := os.Stat(queueDir)
	if err != nil || !info.IsDir() {
		rel, _ := filepath.Rel(root, queueDir)
		return []string{fmt.Sprintf("%s/%s: no queue directory (%s)", lang, topicKey, filepath.ToSlash(rel))}
	}

	remaining := rate
	var messages []string

	files, err := filepath.Glob(filepath.Join(queueDir, "*.md"))
	if err != nil {
		return []string{fmt.Sprintf("%s/%s: %v", lang, topicKey, err)}
	}
	sort.Strings(files)

	for _, queueFile := range files {
		if remaining <= 0 {
			break
		}
		data, err := os.ReadFile(queueFile)
		if err != nil {
			continue
		}
		cards := cardsplit.Split(string(data))
		if len(cards) == 0 {
			continue
		}
		n := remaining
		if n > len(cards) {
			n = len(cards)
		}
		batch := cards[:n]
		rest := cards[n:]
		remaining -= len(batch)

		name := filepath.Base(queueFile)
		target := filepath.Join(cardsDir, name)
		publishedBlock := cardsplit.Join(batch)

		if dryRun {
			rel, _ := filepath.Rel(root, target)
			messages = append(messages, fmt.Sprintf(
				"%s/%s: would publish %d card(s) from %s -> %s",
				lang, topicKey, len(batch), name, filepath.ToSlash(rel),
			))
			continue
		}

		if err := os.MkdirAll(cardsDir, 0o755); err != nil {
			messages = append(messages, fmt.Sprintf("%s/%s: %v", lang, topicKey, err))
			continue
		}
		var merged string
		if _, err := os.Stat(target); err == nil {
			existing, err := os.ReadFile(target)
			if err != nil {
				messages = append(messages, fmt.Sprintf("%s/%s: %v", lang, topicKey, err))
				continue
			}
			merged = strings.TrimRight(string(existing), " \t\n") + "\n\n---\n\n" + strings.TrimRight(publishedBlock, " \t\n") + "\n"
		} else {
			merged = publishedBlock
		}
		if err := os.WriteFile(target, []byte(merged), 0o644); err != nil {
			messages = append(messages, fmt.Sprintf("%s/%s: %v", lang, topicKey, err))
			continue
		}
		if len(rest) > 0 {
			_ = os.WriteFile(queueFile, []byte(cardsplit.Join(rest)), 0o644)
		} else {
			_ = os.Remove(queueFile)
		}
		rel, _ := filepath.Rel(root, target)
		messages = append(messages, fmt.Sprintf(
			"%s/%s: published %d card(s) from %s -> %s",
			lang, topicKey, len(batch), name, filepath.ToSlash(rel),
		))
	}

	if rate > 0 && len(messages) == 0 {
		messages = append(messages, fmt.Sprintf("%s/%s: queue empty, nothing to publish", lang, topicKey))
	}
	return messages
}
