# Cursor Automation: generate draft cards

Runbook for the scheduled Cursor agent that writes kid flashcards into `draft/`.

## Goal

Each run: use your model to create new hashcards markdown for **one topic**, write files under `draft/{Topic}/`, validate, commit, and push to `main`. Do **not** move cards to `queue/` or `cards/` — a human reviews `draft/` first.

## Before writing

1. Read `publish.yaml` — use the `generate` section and each topic's `generate.enabled`, `generate.cards`, `age_bands`, and `subtopics`.
2. Read `templates/card-style.md` for tone and format.
3. Read `templates/topic-seeds/{topic}.md` for ideas.
4. Scan existing content in `cards/`, `queue/`, and `draft/` for the chosen topic so you **do not duplicate** questions or clozes.

## Pick one topic per run

Rotate through topics where `generate.enabled: true`. Prefer the topic with the **fewest files** in `draft/` + `queue/` combined, or the one not generated most recently (use git log on `draft/` if needed).

Within the topic:

- Pick an **age band** from the topic's `age_bands` (rotate across runs).
- Pick a **subtopic** from `subtopics` that is underrepresented in existing files.
- Write exactly `generate.cards` cards (or the topic override).

## Card format (hashcards)

```markdown
# {Subtopic} — {age band label}

<!-- age: early | topic: geography | subtopic: continents -->

Q: Simple question?
A: Short kid-friendly answer.

---

C: Earth has [seven] continents.
```

Rules:

- Separate cards with `---` on its own line.
- Use `Q:` / `A:` pairs or `C:` cloze lines only.
- One idea per card; short answers (see `templates/card-style.md`).
- Filename: `draft/{Topic}/{subtopic}_{age_band}.md` or append to an existing draft file for that subtopic/band if it already exists and has room.

Topic folder names use PascalCase: `Geography`, `Space`, `Animals`, `Sports`, `Science`.

## After writing

```sh
python3 scripts/validate.py
```

Fix any validation errors before committing.

## Commit and push

Only commit files under `draft/` (and this runbook if you changed it).

```sh
git add draft/
git commit -m "draft({topic}): add {n} cards ({subtopic}, {age_band})"
git push origin main
```

If nothing new was written (topic saturated, duplicates only), exit without an empty commit.

## Do not

- Edit `cards/` or `queue/` in this automation.
- Delete existing draft files.
- Use external APIs; write cards with your own model.
- Commit `.venv/`, `*.db`, or secrets.

## Full pipeline (for humans)

```
draft/  →  queue/  →  cards/
  ↑           ↑          ↑
 this      you move    publish.yml
automation   manually   (GitHub Action)
```
