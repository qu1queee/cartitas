#!/usr/bin/env python3
"""Generate draft kid flashcards using an LLM and topic seeds."""

from __future__ import annotations

import argparse
import os
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
STYLE = ROOT / "templates" / "card-style.md"
TOPICS = ["geography", "space", "animals", "sports", "science"]


def topic_dir_name(topic_key: str) -> str:
    return topic_key.replace("_", " ").title().replace(" ", "")


def load_languages() -> list[str]:
    path = ROOT / "languages.yaml"
    with path.open(encoding="utf-8") as f:
        cfg = yaml.safe_load(f)
    return list(cfg.get("languages", {}).keys())


def load_seed(topic: str) -> str:
    seed_path = ROOT / "templates" / "topic-seeds" / f"{topic}.md"
    if not seed_path.is_file():
        raise FileNotFoundError(f"No seed file: {seed_path}")
    return seed_path.read_text(encoding="utf-8")


def lang_label(code: str) -> str:
    return {"en": "English", "es": "Spanish", "de": "German"}.get(code, code)


def build_prompt(
    topic: str,
    age_band: str,
    count: int,
    lang: str,
    *,
    subtopic: str,
    en_reference: str | None = None,
) -> str:
    style = STYLE.read_text(encoding="utf-8") if STYLE.is_file() else ""
    seed = load_seed(topic)
    reference_block = ""
    if en_reference and lang != "en":
        reference_block = f"""
The English version (same facts, same card count — adapt naturally to {lang_label(lang)}):

{en_reference}
"""
    return f"""You write flashcards for kids using hashcards markdown syntax.

Rules:
- Output ONLY markdown cards, no preamble or explanation.
- Separate each card with --- on its own line.
- Use Q:/A: pairs or C: cloze lines.
- Write entirely in {lang_label(lang)} (language code: {lang}).
- Age band: {age_band}
- Topic: {topic}
- Subtopic: {subtopic}
- Write exactly {count} cards.
- Same facts and card order as the English set when translating.
{reference_block}
Follow this style guide:

{style}

Topic seeds and ideas:

{seed}

Start with:
# {subtopic.replace("_", " ").title()} — {age_band}

<!-- age: {age_band} | lang: {lang} | topic: {topic} | subtopic: {subtopic} -->
"""


def generate_with_openai(prompt: str) -> str:
    try:
        from openai import OpenAI
    except ImportError as exc:
        raise SystemExit("Install openai: pip install openai") from exc

    api_key = os.environ.get("OPENAI_API_KEY")
    if not api_key:
        raise SystemExit("OPENAI_API_KEY is not set")

    client = OpenAI(api_key=api_key)
    response = client.chat.completions.create(
        model=os.environ.get("OPENAI_MODEL", "gpt-4o-mini"),
        messages=[{"role": "user", "content": prompt}],
    )
    text = (response.choices[0].message.content or "").strip()
    if not text:
        raise RuntimeError("Empty response from model")
    return text + "\n"


def write_draft(
    topic: str,
    age_band: str,
    lang: str,
    subtopic: str,
    content: str,
) -> Path:
    topic_name = topic_dir_name(topic)
    out_dir = ROOT / "draft" / lang / topic_name
    out_dir.mkdir(parents=True, exist_ok=True)
    path = out_dir / f"{subtopic}_{age_band}.md"
    path.write_text(content, encoding="utf-8")
    return path


def generate_one(
    topic: str,
    age_band: str,
    lang: str,
    count: int,
    subtopic: str,
    en_reference: str | None = None,
) -> Path:
    prompt = build_prompt(
        topic, age_band, count, lang, subtopic=subtopic, en_reference=en_reference
    )
    content = generate_with_openai(prompt)
    if "Q:" not in content and "C:" not in content:
        raise RuntimeError(f"Model output missing cards for lang={lang}")
    path = write_draft(topic, age_band, lang, subtopic, content)
    print(f"Wrote draft: {path.relative_to(ROOT)}")
    return path


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate draft cartitas cards.")
    parser.add_argument("--topic", required=True, choices=TOPICS)
    parser.add_argument("--subtopic", required=True, help="Subtopic slug, e.g. continents")
    parser.add_argument("--age-band", default="early", choices=["early", "middle", "teen"])
    parser.add_argument("--lang", choices=["en", "es", "de"], help="Single language (ignored if --all-langs)")
    parser.add_argument("--all-langs", action="store_true", help="Generate en, es, and de together")
    parser.add_argument("--count", type=int, default=5)
    args = parser.parse_args()

    langs = load_languages() if args.all_langs else [args.lang or "en"]

    en_content: str | None = None
    for lang in langs:
        if lang == "en":
            path = generate_one(
                args.topic, args.age_band, lang, args.count, args.subtopic
            )
            en_content = path.read_text(encoding="utf-8")
        else:
            generate_one(
                args.topic,
                args.age_band,
                lang,
                args.count,
                args.subtopic,
                en_reference=en_content,
            )

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
