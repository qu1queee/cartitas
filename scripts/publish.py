#!/usr/bin/env python3
"""Move approved cards from queue/{lang}/{Topic}/ into cards/{lang}/{Topic}/."""

from __future__ import annotations

import argparse
import re
import sys
from datetime import date
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
CARD_SEP = re.compile(r"^---\s*$", re.MULTILINE)


def load_config(path: Path) -> dict:
    with path.open(encoding="utf-8") as f:
        return yaml.safe_load(f)


def load_languages(path: Path) -> list[str]:
    cfg = load_config(path)
    return list(cfg.get("languages", {}).keys())


def split_cards(text: str) -> list[str]:
    text = text.strip()
    if not text:
        return []
    parts = CARD_SEP.split(text)
    return [p.strip() for p in parts if p.strip()]


def join_cards(cards: list[str]) -> str:
    if not cards:
        return ""
    return "\n\n---\n\n".join(cards) + "\n"


def topic_dir_name(topic_key: str) -> str:
    return topic_key.replace("_", " ").title().replace(" ", "")


def publish_topic(
    topic_key: str,
    topic_cfg: dict,
    defaults: dict,
    lang: str,
    *,
    dry_run: bool,
) -> list[str]:
    rate = topic_cfg.get("rate", defaults.get("cards_per_publish", 2))
    topic_name = topic_dir_name(topic_key)
    queue_dir = ROOT / "queue" / lang / topic_name
    cards_dir = ROOT / "cards" / lang / topic_name

    if not queue_dir.is_dir():
        return [f"{lang}/{topic_key}: no queue directory ({queue_dir.relative_to(ROOT)})"]

    remaining = rate
    messages: list[str] = []

    for queue_file in sorted(queue_dir.glob("*.md")):
        if remaining <= 0:
            break

        source = queue_file.read_text(encoding="utf-8")
        cards = split_cards(source)
        if not cards:
            continue

        batch = cards[:remaining]
        rest = cards[remaining:]
        remaining -= len(batch)

        target = cards_dir / queue_file.name
        published_block = join_cards(batch)

        if dry_run:
            messages.append(
                f"{lang}/{topic_key}: would publish {len(batch)} card(s) "
                f"from {queue_file.name} -> {target.relative_to(ROOT)}"
            )
            continue

        cards_dir.mkdir(parents=True, exist_ok=True)
        if target.exists():
            existing = target.read_text(encoding="utf-8").rstrip()
            merged = existing + "\n\n---\n\n" + published_block.rstrip() + "\n"
        else:
            merged = published_block

        target.write_text(merged, encoding="utf-8")

        if rest:
            queue_file.write_text(join_cards(rest), encoding="utf-8")
        else:
            queue_file.unlink()

        messages.append(
            f"{lang}/{topic_key}: published {len(batch)} card(s) "
            f"from {queue_file.name} -> {target.relative_to(ROOT)}"
        )

    if rate > 0 and not messages:
        messages.append(f"{lang}/{topic_key}: queue empty, nothing to publish")

    return messages


def main() -> int:
    parser = argparse.ArgumentParser(description="Publish queued cartitas cards.")
    parser.add_argument("--config", type=Path, default=ROOT / "publish.yaml")
    parser.add_argument("--languages-config", type=Path, default=ROOT / "languages.yaml")
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--lang", action="append", dest="langs", help="Language code (en, es, de). Repeatable.")
    parser.add_argument("--topic", action="append", dest="topics", help="Topic key. Repeatable.")
    args = parser.parse_args()

    cfg = load_config(args.config)
    defaults = cfg.get("defaults", {})
    topics = cfg.get("topics", {})
    supported = load_languages(args.languages_config)
    selected_langs = args.langs or supported

    for lang in selected_langs:
        if lang not in supported:
            print(f"Unknown language: {lang}", file=sys.stderr)
            return 1

    if not topics:
        print("No topics defined in publish.yaml", file=sys.stderr)
        return 1

    selected_topics = args.topics or list(topics.keys())
    any_published = False

    for lang in selected_langs:
        for topic_key in selected_topics:
            if topic_key not in topics:
                print(f"Unknown topic: {topic_key}", file=sys.stderr)
                return 1
            topic_cfg = topics[topic_key]
            if not topic_cfg.get("enabled", True):
                print(f"{lang}/{topic_key}: disabled, skipping")
                continue

            for msg in publish_topic(
                topic_key, topic_cfg, defaults, lang, dry_run=args.dry_run
            ):
                print(msg)
                if "published" in msg or "would publish" in msg:
                    any_published = True

    if not args.dry_run and any_published:
        print(f"\nPublish complete ({date.today().isoformat()}). Run: hashcards check cards/")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
