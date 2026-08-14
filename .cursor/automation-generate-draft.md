# Cursor Automation: generate draft cards

Runbook for the scheduled Cursor agent that writes kid flashcards into `draft/` and **updates one open PR** (never pushes to `main` directly).

## Goal

Each run: create new hashcards markdown for **one topic in one language**, write under `draft/{lang}/{Topic}/`, validate, commit to branch `automation/draft-cards`, push, and **ensure an open PR** exists targeting `main`.

Do **not** move cards to `queue/` or `cards/` — a human reviews the PR first.

## Languages (required)

Read [languages.yaml](../languages.yaml): `en`, `es`, `de`.

Every run targets **exactly one language**. Rotate across runs:

```
en → es → de → en → …
```

Also rotate **topic** (geography, space, animals, sports, science). Prefer the `(lang, topic)` pair with the fewest files in `draft/{lang}/` + `queue/{lang}/` combined.

When writing `es` or `de`, match the age band and subtopic of existing `en` content when possible so facts stay aligned across languages.

## Before writing

1. Read `languages.yaml` and `publish.yaml` (`generate` section).
2. Read `templates/card-style.md` and `templates/topic-seeds/{topic}.md`.
3. Scan `cards/{lang}/`, `queue/{lang}/`, and `draft/{lang}/` for duplicates.
4. Check for an existing open PR: `gh pr list --head automation/draft-cards --state open`

## Pick content for this run

- **Language:** one of `en`, `es`, `de` (rotate).
- **Topic:** one with `generate.enabled: true`.
- **Age band:** from topic's `age_bands` (rotate).
- **Subtopic:** underrepresented in that lang/topic.
- **Count:** `generate.cards` from publish.yaml (topic override if set).
- Write in the **target language only** — natural kid-friendly prose, not literal translationese.

## Card format

```markdown
# {Subtopic} — {age band label}

<!-- age: early | lang: es | topic: geography | subtopic: continents -->

Q: ¿Cuántos continentes hay?
A: Siete: África, Antártida, Asia, Europa, América del Norte, Oceanía y América del Sur.

---

C: La Tierra tiene [siete] continentes.
```

- Path: `draft/{lang}/{Topic}/{subtopic}_{age_band}.md`
- Topic folders: `Geography`, `Space`, `Animals`, `Sports`, `Science` (PascalCase).

## Validate

```sh
python3 scripts/validate.py
```

Fix errors before committing.

## Commit via PR branch (do not push to main)

Use the helper script — it rebases on `main`, pushes `automation/draft-cards`, and opens a PR if none exists:

```sh
python3 scripts/draft_pr.py \
  --message "draft(es/geography): add 5 cards (continents, early)" \
  --pr-note "Language: es | Topic: geography | Band: early | Subtopic: continents"
```

Dry run first if unsure:

```sh
python3 scripts/draft_pr.py --dry-run --message "draft(es/geography): ..."
```

### Manual equivalent (only if script fails)

```sh
git fetch origin main
git checkout automation/draft-cards 2>/dev/null || git checkout -b automation/draft-cards origin/main
git pull --rebase origin automation/draft-cards 2>/dev/null || true
git rebase origin/main
git add draft/
git commit -m "draft({lang}/{topic}): add {n} cards ({subtopic}, {age_band})"
git push -u origin automation/draft-cards
gh pr list --head automation/draft-cards --state open || \
  gh pr create --base main --head automation/draft-cards \
    --title "Draft cards (automation)" \
    --body "Automated draft flashcards. Review draft/{lang}/ then move to queue/{lang}/."
```

## After merge (human)

1. Review the PR on GitHub.
2. Move approved files: `draft/{lang}/` → `queue/{lang}/`.
3. Publish workflow moves `queue/` → `cards/`.
4. Kids drill: `python scripts/drill.py --lang es`

## Do not

- Push to `main` directly.
- Edit `cards/` or `queue/` in this automation.
- Mix languages in one file.
- Close the draft PR after each run — **reuse the same open PR** until merged.

## Pipeline

```
automation → draft/{lang}/ → PR (automation/draft-cards)
                                    ↓ merge + you move files
                              queue/{lang}/ → cards/{lang}/
```
