#!/usr/bin/env python3
"""Generate draft kid flashcards using an LLM and topic seeds."""

from __future__ import annotations

import argparse
import os
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
STYLE = ROOT / "templates" / "card-style.md"


def topic_dir_name(topic_key: str) -> str:
    return topic_key.replace("_", " ").title().replace(" ", "")


def load_seed(topic: str) -> str:
    seed_path = ROOT / "templates" / "topic-seeds" / f"{topic}.md"
    if not seed_path.is_file():
        raise FileNotFoundError(f"No seed file: {seed_path}")
    return seed_path.read_text(encoding="utf-8")


def build_prompt(topic: str, age_band: str, count: int) -> str:
    style = STYLE.read_text(encoding="utf-8") if STYLE.is_file() else ""
    seed = load_seed(topic)
    return f"""You write flashcards for kids using hashcards markdown syntax.

Rules:
- Output ONLY markdown cards, no preamble or explanation.
- Separate each card with --- on its own line.
- Use Q:/A: pairs or C: cloze lines.
- Age band: {age_band}
- Topic: {topic}
- Write exactly {count} cards.
- Follow this style guide:

{style}

Topic seeds and ideas:

{seed}

Start with a single # section header line, then optional metadata comment, then cards.
Example header:
# {{Subtopic}} — {age_band}

<!-- age: {age_band} | topic: {topic} | subtopic: generated -->
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


def write_draft(topic: str, age_band: str, content: str) -> Path:
    topic_name = topic_dir_name(topic)
    out_dir = ROOT / "draft" / topic_name
    out_dir.mkdir(parents=True, exist_ok=True)

    existing = sorted(out_dir.glob(f"*_{age_band}.md"))
    n = len(existing) + 1
    path = out_dir / f"generated_{n}_{age_band}.md"
    path.write_text(content, encoding="utf-8")
    return path


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate draft cartitas cards.")
    parser.add_argument("--topic", required=True, choices=[
        "geography", "space", "animals", "sports", "science"
    ])
    parser.add_argument("--age-band", default="early", choices=["early", "middle", "teen"])
    parser.add_argument("--count", type=int, default=5)
    args = parser.parse_args()

    prompt = build_prompt(args.topic, args.age_band, args.count)
    content = generate_with_openai(prompt)

    if "Q:" not in content and "C:" not in content:
        print("Model output missing cards:", file=sys.stderr)
        print(content, file=sys.stderr)
        return 1

    path = write_draft(args.topic, args.age_band, content)
    print(f"Wrote draft: {path.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
