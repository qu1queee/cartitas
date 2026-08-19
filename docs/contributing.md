# Contributing

Internal notes for writing, reviewing, and publishing cards.

## Layout

| Path | Purpose |
|------|---------|
| `cards/{lang}/{Topic}/` | Published cards, ready to drill |
| `queue/{lang}/{Topic}/` | Cards waiting for auto-publish |
| `draft/{lang}/{Topic}/` | Human work in progress (not used by automation) |
| `languages.yaml` | Supported languages and default |
| `templates/` | Style guide and topic seeds |
| `publish.yaml` | Per-topic publish rates |

Languages: `en` (English), `es` (Español), `de` (Deutsch). See [languages.yaml](../languages.yaml).

Filename: `{subtopic}_{band}.md` inside each language folder.

## Card format

```markdown
Q: What is the biggest planet?
A: Jupiter. It is so big that all the other planets could fit inside it.

---

C: Earth goes around the [Sun] once every year.
```

One language per file. Metadata: `<!-- age: early | lang: en | topic: space | subtopic: planets -->`

Age bands and writing rules: [templates/card-style.md](../templates/card-style.md).

| Band | Ages | Use |
|------|------|-----|
| toddler | 2–4 | Naming, cloze |
| early | 5–7 | Simple Q/A |
| middle | 8–11 | Why and compare |
| teen | 12+ | Deeper facts, still plain language |

## Pipeline

```
you (optional):  draft/{lang}/  →  queue/{lang}/
automation:      queue/{lang}/  (via PR automation/queue-cards)
then:            queue/{lang}/  →  cards/{lang}/
                         (publish workflow)
```

1. **You:** write in `draft/{lang}/`, review in your own PR, then move all three language files together to `queue/{lang}/{Topic}/`.
2. **Automation:** writes straight to `queue/{lang}/` and opens/updates one PR. Merge when you want those files to publish.
3. GitHub Actions (or `go run ./cmd/cartitas publish`) publishes `queue/{lang}/` into `cards/{lang}/`.

### Publish locally

```sh
go run ./cmd/cartitas publish --dry-run
go run ./cmd/cartitas publish --lang es
go run ./cmd/cartitas publish
```

### GitHub Actions

| Workflow | Trigger | What it does |
|----------|---------|--------------|
| [Validate](../.github/workflows/validate.yml) | Push / PR | Syntax + trilingual queue check + hashcards |
| [Publish](../.github/workflows/publish.yml) | Daily 08:00 UTC + manual | Moves `queue/{lang}/` → `cards/{lang}/` |

Queued cards are **not** generated in GitHub Actions. The Cursor Automation writes `queue/` and opens the PR.

## Cursor Automation (queue)

Scheduled agent writes **the same new cards in en, es, and de** each run, then **opens/updates one PR** (`automation/queue-cards`).

1. Runbook: [.cursor/automation-generate-queue.md](../.cursor/automation-generate-queue.md)
2. Each run creates three files: `queue/{en,es,de}/{Topic}/{subtopic}_{band}.md`
3. Helper: `go run ./cmd/cartitas queue-pr --message "queue(geography/continents): add 5 cards x3 langs (early)"` — PR is **ready for review**, not a GitHub draft.

Prefill: [.cursor/automation-prefill.json](../.cursor/automation-prefill.json) — re-save the automation after pulling latest `main`.

## Topics

| Topic | Example decks |
|-------|---------------|
| Geography | continents, oceans |
| Space | solar system, moon |
| Animals | pets, farm, ocean |
| Sports | football, basketball |
| Science | body, plants |

## Validation

```sh
go run ./cmd/cartitas validate
hashcards check cards/
```
