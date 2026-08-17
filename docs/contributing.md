# Contributing

Internal notes for writing, reviewing, and publishing cards.

## Layout

| Path | Purpose |
|------|---------|
| `cards/{lang}/{Topic}/` | Published cards, ready to drill |
| `queue/{lang}/{Topic}/` | Approved cards waiting for auto-publish |
| `draft/{lang}/{Topic}/` | Work in progress |
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
draft/{lang}/  →  queue/{lang}/  →  cards/{lang}/
         (you)           (automated)
```

1. Write or generate cards in `draft/{lang}/`
2. Review and move good ones to `queue/{lang}/{Topic}/`
3. GitHub Actions (or `go run ./cmd/cartitas publish`) publishes into `cards/{lang}/`

After a draft PR merges, move **all three language files together** to `queue/{lang}/`.

### Publish locally

```sh
go run ./cmd/cartitas publish --dry-run
go run ./cmd/cartitas publish --lang es
go run ./cmd/cartitas publish
```

### GitHub Actions

| Workflow | Trigger | What it does |
|----------|---------|--------------|
| [Validate](../.github/workflows/validate.yml) | Push / PR | Syntax + trilingual draft check + hashcards |
| [Publish](../.github/workflows/publish.yml) | Daily 08:00 UTC + manual | Moves `queue/{lang}/` → `cards/{lang}/` |

Drafts are **not** generated in GitHub Actions. The Cursor Automation writes `draft/` and opens the PR.

## Cursor Automation (draft)

Scheduled agent writes **the same new cards in en, es, and de** each run, then **opens/updates one PR** (`automation/draft-cards`).

1. Runbook: [.cursor/automation-generate-draft.md](../.cursor/automation-generate-draft.md)
2. Each run creates three files: `draft/{en,es,de}/{Topic}/{subtopic}_{band}.md`
3. Helper: `go run ./cmd/cartitas draft-pr --message "draft(geography/continents): add 5 cards x3 langs (early)"`

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
