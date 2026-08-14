# Cursor Automation: generate draft cards

Runbook for the scheduled Cursor agent that writes kid flashcards into `draft/`.

## Goal

Each run: use your model to create new hashcards markdown for **one topic in one language**, write files under `draft/{lang}/{Topic}/`, validate, commit, and push to `main`. Do **not** move cards to `queue/` or `cards/` — a human reviews `draft/` first.

## Before writing

1. Read `languages.yaml` — supported codes: `en`, `es`, `de`.
2. Read `publish.yaml` — use the `generate` section and each topic's settings.
3. Read `templates/card-style.md` for tone and format.
4. Read `templates/topic-seeds/{topic}.md` for ideas.
5. Scan `cards/{lang}/`, `queue/{lang}/`, and `draft/{lang}/` for the chosen topic so you **do not duplicate** questions or clozes.

## Pick one language and topic per run

Rotate **language** (`en` → `es` → `de` → …) and **topic** where `generate.enabled: true`. Prefer combinations with the fewest files in `draft/` + `queue/`.

Within the topic:

- Pick an **age band** from the topic's `age_bands` (rotate across runs).
- Pick a **subtopic** from `subtopics` that is underrepresented.
- Write exactly `generate.cards` cards (or the topic override).
- Write in the **target language only** (natural kid-friendly Spanish or German, not word-for-word calques).

## Card format (hashcards)

```markdown
# {Subtopic} — {age band label}

<!-- age: early | lang: es | topic: geography | subtopic: continents -->

Q: Simple question?
A: Short kid-friendly answer.

---

C: La Tierra tiene [siete] continentes.
```

Rules:

- Separate cards with `---` on its own line.
- Use `Q:` / `A:` pairs or `C:` cloze lines only.
- One idea per card; short answers (see `templates/card-style.md`).
- Path: `draft/{lang}/{Topic}/{subtopic}_{age_band}.md`

Topic folder names use PascalCase: `Geography`, `Space`, `Animals`, `Sports`, `Science`.

## After writing

```sh
python3 scripts/validate.py
```

Fix any validation errors before committing.

## Commit and push

Only commit files under `draft/`.

```sh
git add draft/
git commit -m "draft({lang}/{topic}): add {n} cards ({subtopic}, {age_band})"
git push origin main
```

If nothing new was written, exit without an empty commit.

## Do not

- Edit `cards/` or `queue/` in this automation.
- Mix languages in one file.
- Delete existing draft files.
- Use external APIs; write cards with your own model.
- Commit `.venv/`, `*.db`, or secrets.

## Full pipeline (for humans)

```
draft/{lang}/  →  queue/{lang}/  →  cards/{lang}/
      ↑                ↑                  ↑
   this automation   you move         publish.yml
```

Kids drill with: `python scripts/drill.py --lang es`
