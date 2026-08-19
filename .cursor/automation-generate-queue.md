# Cursor Automation: generate queued cards

Runbook for the scheduled Cursor agent that writes kid flashcards into `queue/` and **updates one open PR** (never pushes to `main` directly).

`draft/` is for **human** PRs you review yourself. This agent must **not** write under `draft/`.

## Goal

Each run: create **the same new cards in all three languages** (`en`, `es`, `de`) for **one topic**, validate, commit to branch `automation/queue-cards`, push, and **ensure an open PR** exists.

Do **not** edit `cards/` or `draft/` — merge puts files in `queue/`; publish moves `queue/` → `cards/`.

## Languages (required — all three every run)

Read [languages.yaml](../languages.yaml): `en`, `es`, `de`.

**Every run produces a trilingual set.** Same subtopic, age band, card count, and facts — three separate files:

```
queue/en/{Topic}/{subtopic}_{age_band}.md
queue/es/{Topic}/{subtopic}_{age_band}.md
queue/de/{Topic}/{subtopic}_{age_band}.md
```

Workflow:

1. Write **English** first (`queue/en/...`).
2. Write **Spanish** and **German** with the **same facts and card order** (natural kid-friendly wording, not awkward literal translation).
3. All three files must exist before you commit.

Rotate **topic** each run (geography, space, animals, sports, science). Pick the subtopic most underrepresented across all three languages.

## Before writing

1. Read `languages.yaml` and `publish.yaml` (`generate` section).
2. Read `templates/card-style.md` and `templates/topic-seeds/{topic}.md`.
3. Scan `cards/{lang}/`, `queue/{lang}/`, and `draft/{lang}/` for duplicates in **all three** languages.
4. Check for an existing open PR: `gh pr list --head automation/queue-cards --state open`

## Pick content for this run

- **Topic:** one with `generate.enabled: true`.
- **Subtopic:** from topic's `subtopics`, underrepresented everywhere.
- **Age band:** from topic's `age_bands` (rotate).
- **Count:** `generate.cards` from publish.yaml.
- **Languages:** always `en` + `es` + `de` in the same run.

## Card format

```markdown
# Continents — early

<!-- age: early | lang: en | topic: geography | subtopic: continents -->

Q: How many continents are there?
A: Seven. They are Africa, Antarctica, Asia, Europe, North America, Oceania, and South America.
```

Spanish and German files use the same header/subtopic slug and matching cards with `lang: es` / `lang: de`.

- Path pattern: `queue/{lang}/{Topic}/{subtopic}_{age_band}.md`
- Topic folders: `Geography`, `Space`, `Animals`, `Sports`, `Science`

## Validate (both required)

```sh
go run ./cmd/cartitas validate
go run ./cmd/cartitas validate-trilingual --stage queue
```

`validate-trilingual` **must pass** — it fails if any queue file is missing a matching `en`, `es`, or `de` sibling with the same path under `queue/{lang}/`.

Do not commit if trilingual validation fails.

## Commit via PR branch (do not push to main)

```sh
go run ./cmd/cartitas queue-pr \
  --message "queue(geography/continents): add 5 cards x3 langs (early)" \
  --pr-note "Topic: geography | Subtopic: continents | Band: early | Langs: en, es, de"
```

The PR on `automation/queue-cards` must be **ready for review**, not a GitHub draft. `queue-pr` creates it without `--draft` and runs `gh pr ready` if it is still a draft. If Cursor opened a draft anyway:

```sh
gh pr ready --repo qu1queee/cartitas automation/queue-cards
```

## After merge

1. Merged files already sit in `queue/{en,es,de}/{Topic}/`.
2. Publish workflow moves `queue/` → `cards/` (per language).
3. Kids drill: `go run ./cmd/cartitas drill --lang es`

## Do not

- Push to `main` directly.
- Write under `draft/` (that path is for human-reviewed PRs).
- Create cards in only one language per run.
- Run `gh pr create` on `cursor/*` branches — **always** use `go run ./cmd/cartitas queue-pr` (branch `automation/queue-cards` only).
- Leave the queue PR as a GitHub draft (`gh pr create --draft`). It must stay **ready for review**.
- Skip `validate-trilingual`.
- Mix languages in one file.
- Use different filenames across languages for the same card set.

## Pipeline

```
you (optional)  → draft/{lang}/ → you review, then move to queue
automation      → queue/{en,es,de}/ → PR (automation/queue-cards) → merge
                  queue/{lang}/ → cards/{lang}/
```
