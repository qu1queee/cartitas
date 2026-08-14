# Cartitas card style guide

Cartitas cards are short, friendly facts for kids from about age 2 through adolescence.
They use the same syntax as [hashcards](https://github.com/SimonPersson/hashcards).

## Card types

### Question and answer (most common)

```markdown
Q: What is the biggest planet?
A: Jupiter. It is so big that all the other planets could fit inside it.
```

### Cloze (fill in the blank)

```markdown
C: Earth goes around the [Sun] once every year.
```

Use cloze for names, colors, numbers, and simple facts.

## Age bands

| Band | Ages | Answer length | Tone |
|------|------|---------------|------|
| toddler | 2–4 | One short sentence | Naming, colors, sounds |
| early | 5–7 | 1–2 sentences | Simple what/where/who |
| middle | 8–11 | 2–3 sentences | Why and compare |
| teen | 12+ | Up to 4 sentences | Deeper but still plain |

## Writing rules

1. Use words a kid at that age would know.
2. One idea per card.
3. Prefer concrete examples ("A dog says woof") over abstract definitions.
4. Avoid jargon; if you must use a big word, explain it simply.
5. Answers should stand alone (no "see above").
6. Fun is good; jokes are fine if the fact stays clear.

## Optional metadata (ignored by hashcards)

```markdown
<!-- age: early | lang: en | topic: space | subtopic: planets -->
```

Use `lang: en`, `lang: es`, or `lang: de`. **One language per file.**

When adding new content, always create **matching files in all three languages** with the same `{subtopic}_{band}.md` filename:

```
draft/en/Geography/continents_early.md
draft/es/Geography/continents_early.md
draft/de/Geography/continents_early.md
```

## Languages

| Code | Language |
|------|----------|
| en | English |
| es | Español |
| de | Deutsch |

Content paths include the language code first:

- `cards/{lang}/{Topic}/{subtopic}_{band}.md`
- `queue/{lang}/{Topic}/`
- `draft/{lang}/{Topic}/`

Drill one language: `python scripts/drill.py --lang es`

## File layout

- `cards/{lang}/{Topic}/{subtopic}_{band}.md` — published, ready to drill
- `queue/{lang}/{Topic}/` — approved, waiting to publish
- `draft/{lang}/{Topic}/` — work in progress

Separate cards with `---` on its own line.

## Section headers

Use `# Title` at the top of a file for humans. hashcards does not turn headers into cards.

## Example file

```markdown
# Planets — early (5–7)

<!-- age: early | lang: en | topic: space | subtopic: planets -->

Q: What is the biggest planet?
A: Jupiter. It is so big that all the other planets could fit inside it.

---

C: Earth goes around the [Sun] once every year.

---

Q: Which planet is closest to the Sun?
A: Mercury.
```
