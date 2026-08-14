# Cartitas

A kid-friendly flashcard knowledge base — like [cartas](https://github.com/qu1queee/cartas), but for children from toddler years through adolescence.

Topics start with **geography**, **space**, **animals**, **sports**, and **science**. Cards are available in **English**, **Español**, and **Deutsch**. [hashcards](https://github.com/SimonPersson/hashcards) handles spaced repetition locally.

## Quick start

**Install hashcards** (once):

```sh
cargo install hashcards --locked
```

**Drill in a language:**

```sh
python3 scripts/drill.py --lang en          # English (default)
python3 scripts/drill.py --lang es          # Español
python3 scripts/drill.py --lang de --topic Animals
python3 scripts/drill.py --list               # show languages
```

Or directly with hashcards:

```sh
hashcards drill cards/es/ --new-card-limit 5 --answer-controls binary
hashcards drill cards/de/Animals/
```

The drill UI runs at `http://127.0.0.1:8000`.

## Content layout

| Path | Purpose |
|------|---------|
| `cards/{lang}/{Topic}/` | Published cards, ready to drill |
| `queue/{lang}/{Topic}/` | Approved cards waiting for auto-publish |
| `draft/{lang}/{Topic}/` | Work in progress |
| `languages.yaml` | Supported languages and default |
| `templates/` | Style guide and topic seeds |
| `publish.yaml` | Per-topic publish rates |

Languages: `en` (English), `es` (Español), `de` (Deutsch). See [languages.yaml](languages.yaml).

### Card format

```markdown
Q: What is the biggest planet?
A: Jupiter. It is so big that all the other planets could fit inside it.

---

C: Earth goes around the [Sun] once every year.
```

One language per file. Metadata: `<!-- age: early | lang: en | topic: space | subtopic: planets -->`

See [templates/card-style.md](templates/card-style.md) for age bands and writing rules.

## Publishing workflow

```
draft/{lang}/  →  queue/{lang}/  →  cards/{lang}/
         (you)           (automated)
```

1. Write or generate cards in `draft/{lang}/`
2. Review and move good ones to `queue/{lang}/{Topic}/`
3. GitHub Actions (or `scripts/publish.py`) publishes into `cards/{lang}/`

### Publish locally

```sh
python3 -m venv .venv && source .venv/bin/activate
pip install -r scripts/requirements.txt

python scripts/publish.py --dry-run
python scripts/publish.py --lang es          # one language
python scripts/publish.py                     # all languages
```

### GitHub Actions

| Workflow | Trigger | What it does |
|----------|---------|--------------|
| [Publish](.github/workflows/publish.yml) | Daily 08:00 UTC + manual | Moves queue → cards, commits |
| [Validate](.github/workflows/validate.yml) | Push / PR | Syntax + hashcards check |
| [Generate](.github/workflows/generate.yml) | Manual | OpenAI drafts → PR (optional) |

### Cursor Automation (draft)

Scheduled agent writes to `draft/{lang}/` — see [.cursor/automation-generate-draft.md](.cursor/automation-generate-draft.md).

## Topics

Each topic has decks in `cards/{lang}/{Topic}/` for every supported language.

| Topic | Example decks |
|-------|---------------|
| Geography | continents, oceans |
| Space | solar system, moon |
| Animals | pets, farm, ocean |
| Sports | football, basketball |
| Science | body, plants |

## Age bands

| Band | Ages | Use |
|------|------|-----|
| toddler | 2–4 | Naming, cloze |
| early | 5–7 | Simple Q/A |
| middle | 8–11 | Why and compare |
| teen | 12+ | Deeper facts, still plain language |

Filename: `{subtopic}_{band}.md` inside each language folder.

## Daily routine (suggested)

1. Each kid drills in their language: `python scripts/drill.py --lang es --new-card-limit 5`
2. Review `draft/` or add cards to `queue/{lang}/`
3. Publish workflow releases new cards overnight

## Validation

```sh
python scripts/validate.py
hashcards check cards/
```

## Related

- [cartas](https://github.com/qu1queee/cartas) — adult technical flashcards (same format)
- [hashcards](https://github.com/SimonPersson/hashcards) — spaced repetition engine
