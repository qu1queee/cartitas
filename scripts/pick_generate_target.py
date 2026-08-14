#!/usr/bin/env python3
"""Pick the next topic/subtopic/age_band to generate (trilingual gap)."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
TOPIC_ORDER = ["geography", "space", "animals", "sports", "science"]


def topic_dir_name(topic_key: str) -> str:
    return topic_key.replace("_", " ").title().replace(" ", "")


def load_publish_config() -> dict:
    with (ROOT / "publish.yaml").open(encoding="utf-8") as f:
        return yaml.safe_load(f)


def file_exists(base: str, lang: str, topic_name: str, filename: str) -> bool:
    return (ROOT / base / lang / topic_name / filename).is_file()


def lang_coverage(topic: str, subtopic: str, band: str) -> int:
    """Count how many langs have this deck in cards, queue, or draft (0-3)."""
    topic_name = topic_dir_name(topic)
    filename = f"{subtopic}_{band}.md"
    count = 0
    for lang in ("en", "es", "de"):
        if any(
            file_exists(base, lang, topic_name, filename)
            for base in ("cards", "queue", "draft")
        ):
            count += 1
    return count


def pick_target(cfg: dict) -> dict:
    topics = cfg.get("topics", {})
    best = None
    best_score = (999, 999)  # (coverage, topic_index) lower wins

    for topic_key in TOPIC_ORDER:
        topic_cfg = topics.get(topic_key, {})
        gen = topic_cfg.get("generate", {})
        if not topic_cfg.get("enabled", True) or not gen.get("enabled", True):
            continue

        subtopics = topic_cfg.get("subtopics", [])
        bands = topic_cfg.get("age_bands", ["early"])
        cards = gen.get("cards", cfg.get("generate", {}).get("defaults", {}).get("cards", 5))
        topic_idx = TOPIC_ORDER.index(topic_key)

        for subtopic in subtopics:
            for band in bands:
                coverage = lang_coverage(topic_key, subtopic, band)
                score = (coverage, topic_idx)
                if score < best_score:
                    best_score = score
                    best = {
                        "topic": topic_key,
                        "subtopic": subtopic,
                        "age_band": band,
                        "count": cards,
                        "coverage": coverage,
                    }

    if best is None:
        raise SystemExit("No generate target found in publish.yaml")
    return best


def main() -> int:
    parser = argparse.ArgumentParser(description="Pick next trilingual generate target.")
    parser.add_argument(
        "--github-output",
        type=Path,
        help="Append topic/subtopic/age_band/count to GitHub Actions output file",
    )
    parser.add_argument(
        "--export",
        action="store_true",
        help="Print shell exports: topic=... subtopic=...",
    )
    args = parser.parse_args()

    target = pick_target(load_publish_config())

    if args.github_output:
        lines = [
            f"topic={target['topic']}\n",
            f"subtopic={target['subtopic']}\n",
            f"age_band={target['age_band']}\n",
            f"count={target['count']}\n",
        ]
        with args.github_output.open("a", encoding="utf-8") as f:
            f.writelines(lines)

    if args.export:
        for key, val in target.items():
            print(f"{key}={val}")
    else:
        print(
            f"topic={target['topic']} subtopic={target['subtopic']} "
            f"age_band={target['age_band']} count={target['count']} "
            f"(lang coverage {target['coverage']}/3)"
        )

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
