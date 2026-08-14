# Cartitas

A kid-friendly flashcard knowledge base — like [cartas](https://github.com/qu1queee/cartas), but for children from toddler years through adolescence.

Topics start with **geography**, **space**, **animals**, **sports**, and **science**. Cards use simple language and short answers. [hashcards](https://github.com/SimonPersson/hashcards) handles spaced repetition locally.

## Quick start

**Install hashcards** (once):

```sh
cargo install hashcards --locked
```

**Drill published cards:**

```sh
hashcards drill cards/ --new-card-limit 5 --answer-controls binary
```

**Drill one topic:**

```sh
hashcards drill cards/Animals/
hashcards drill cards/ --from-deck pets_early
```

The drill UI runs at `http://127.0.0.1:8000`.

## Content layout

| Path | Purpose |
|------|---------|
| `cards/{Topic}/` | Published cards, ready to drill |
| `queue/{Topic}/` | Approved cards waiting for auto-publish |
| `draft/{Topic}/` | Work in progress (review before queueing) |
| `templates/` | Style guide and topic seeds |
| `publish.yaml` | Per-topic publish rates and schedules |

### Card format

```markdown
Q: What is the biggest planet?
A: Jupiter. It is so big that all the other planets could fit inside it.

---

C: Earth goes around the [Sun] once every year.
```

See [templates/card-style.md](templates/card-style.md) for age bands and writing rules.

## Publishing workflow

Cards move through three stages:

```
draft  →  queue  →  cards/
         (you)     (automated)
```

1. Write or generate cards in `draft/`
2. Review and move good ones to `queue/{Topic}/`
3. GitHub Actions (or local script) publishes N cards per topic into `cards/`

### Configure rates

Edit [publish.yaml](publish.yaml):

```yaml
topics:
  geography:
    enabled: true
    rate: 2    # cards per publish run
```

### Publish locally

```sh
python3 -m venv .venv
source .venv/bin/activate
pip install -r scripts/requirements.txt

python scripts/publish.py --dry-run   # preview
python scripts/publish.py             # publish now
python scripts/publish.py --topic geography
```

### GitHub Actions

| Workflow | Trigger | What it does |
|----------|---------|--------------|
| [Publish](.github/workflows/publish.yml) | Daily 08:00 UTC + manual | Moves queue → cards, commits |
| [Validate](.github/workflows/validate.yml) | Push / PR | Syntax + hashcards check |
| [Generate](.github/workflows/generate.yml) | Manual | OpenAI drafts → PR in `draft/` (optional) |

### Cursor Automation (recommended for draft)

Use a **scheduled Cursor Automation** so a cloud agent writes cards into `draft/` with Cursor’s model — no OpenAI API key in GitHub.

1. Follow the runbook: [.cursor/automation-generate-draft.md](.cursor/automation-generate-draft.md)
2. Schedule: daily **06:00 UTC** (see `generate.schedule` in [publish.yaml](publish.yaml))
3. Repo: `qu1queee/cartitas`, branch `main`
4. After each run: review `draft/`, move good cards to `queue/`

Rates per topic are in `publish.yaml` under each topic’s `generate.cards`.

## Topics

| Topic | Published decks | Queued |
|-------|-----------------|--------|
| [Geography](cards/Geography/) | continents, oceans | countries, continents (middle) |
| [Space](cards/Space/) | solar system, moon | stars, solar system (teen) |
| [Animals](cards/Animals/) | pets, farm | ocean, pets (middle) |
| [Sports](cards/Sports/) | football | basketball, olympics |
| [Science](cards/Science/) | body, plants | weather, body (middle) |

## Age bands

| Band | Ages | Use |
|------|------|-----|
| toddler | 2–4 | Naming, cloze |
| early | 5–7 | Simple Q/A |
| middle | 8–11 | Why and compare |
| teen | 12+ | Deeper facts, still plain language |

Filename suffix: `{subtopic}_{band}.md` (e.g. `planets_early.md`).

## Daily routine (suggested)

1. Kids drill 5–10 cards: `hashcards drill cards/ --new-card-limit 5 --answer-controls binary`
2. You review `draft/` or add cards straight to `queue/`
3. Publish workflow releases new cards overnight

## Validation

```sh
python scripts/validate.py
hashcards check cards/
```

## Related

- [cartas](https://github.com/qu1queee/cartas) — adult technical flashcards (same format)
- [hashcards](https://github.com/SimonPersson/hashcards) — spaced repetition engine
