# Cursor Automation: generate draft cards

Runbook for the scheduled Cursor agent that writes kid flashcards into `draft/` and **updates one open PR** (never pushes to `main` directly).

## Goal

Each run: create **the same new cards in all three languages** (`en`, `es`, `de`) for **one topic**, validate, commit to branch `automation/draft-cards`, push, and **ensure an open PR** exists.

Do **not** move cards to `queue/` or `cards/` — a human reviews the PR first.

## Languages (required — all three every run)

Read [languages.yaml](../languages.yaml): `en`, `es`, `de`.

**Every run produces a trilingual set.** Same subtopic, age band, card count, and facts — three separate files:

```
draft/en/{Topic}/{subtopic}_{age_band}.md
draft/es/{Topic}/{subtopic}_{age_band}.md
draft/de/{Topic}/{subtopic}_{age_band}.md
```

Workflow:

1. Write **English** first (`draft/en/...`).
2. Write **Spanish** and **German** with the **same facts and card order** (natural kid-friendly wording, not awkward literal translation).
3. All three files must exist before you commit.

Rotate **topic** each run (geography, space, animals, sports, science). Pick the subtopic most underrepresented across all three languages.

## Before writing

1. Read `languages.yaml` and `publish.yaml` (`generate` section).
2. Read `templates/card-style.md` and `templates/topic-seeds/{topic}.md`.
3. Scan `cards/{lang}/`, `queue/{lang}/`, and `draft/{lang}/` for duplicates in **all three** languages.
4. Check for an existing open PR: `gh pr list --head automation/draft-cards --state open`

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

- Path pattern: `draft/{lang}/{Topic}/{subtopic}_{age_band}.md`
- Topic folders: `Geography`, `Space`, `Animals`, `Sports`, `Science`

## Validate

```sh
python3 scripts/validate.py
```

All three language files must pass before committing.

## Commit via PR branch (do not push to main)

```sh
python3 scripts/draft_pr.py \
  --message "draft(geography/continents): add 5 cards x3 langs (early)" \
  --pr-note "Topic: geography | Subtopic: continents | Band: early | Langs: en, es, de"
```

## After merge (human)

1. Review all three files in the PR.
2. Move approved sets: `draft/{lang}/` → `queue/{lang}/` for each language.
3. Publish workflow moves `queue/` → `cards/`.
4. Kids drill: `python scripts/drill.py --lang es`

## Do not

- Push to `main` directly.
- Create cards in only one language per run.
- Edit `cards/` or `queue/` in this automation.
- Mix languages in one file.
- Use different filenames across languages for the same card set.

## Pipeline

```
automation → draft/{en,es,de}/ → PR (automation/draft-cards)
                                        ↓ merge + you move files
                              queue/{lang}/ → cards/{lang}/
```
