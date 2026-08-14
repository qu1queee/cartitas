#!/usr/bin/env python3
"""Drill cartitas cards in a chosen language."""

from __future__ import annotations

import argparse
import shutil
import subprocess
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent


def load_languages(path: Path) -> dict:
    with path.open(encoding="utf-8") as f:
        return yaml.safe_load(f)


def main() -> int:
    cfg = load_languages(ROOT / "languages.yaml")
    supported = list(cfg.get("languages", {}).keys())
    default = cfg.get("default", "en")

    parser = argparse.ArgumentParser(description="Drill cartitas in one language.")
    parser.add_argument(
        "--lang",
        choices=supported,
        default=default,
        help=f"Language code (default: {default})",
    )
    parser.add_argument(
        "--topic",
        help="Topic folder, e.g. Animals, Geography, Space",
    )
    parser.add_argument(
        "--new-card-limit",
        type=int,
        default=5,
        help="Max new cards per session (default: 5)",
    )
    parser.add_argument(
        "--list",
        action="store_true",
        help="List supported languages and exit",
    )
    args = parser.parse_args()

    if args.list:
        for code, meta in cfg.get("languages", {}).items():
            mark = " (default)" if code == default else ""
            print(f"{code}: {meta.get('name', code)}{mark}")
        return 0

    if shutil.which("hashcards") is None:
        print("hashcards not found. Install: cargo install hashcards --locked", file=sys.stderr)
        return 1

    cards_dir = ROOT / "cards" / args.lang
    if not cards_dir.is_dir():
        print(f"No cards for language '{args.lang}' at {cards_dir}", file=sys.stderr)
        return 1

    drill_path = cards_dir / args.topic if args.topic else cards_dir
    if not drill_path.is_dir():
        print(f"Path not found: {drill_path}", file=sys.stderr)
        return 1

    lang_name = cfg["languages"][args.lang].get("name", args.lang)
    print(f"Drilling {lang_name} ({args.lang}): {drill_path.relative_to(ROOT)}")

    cmd = [
        "hashcards",
        "drill",
        str(drill_path),
        "--new-card-limit",
        str(args.new_card_limit),
        "--answer-controls",
        "binary",
    ]
    return subprocess.call(cmd)


if __name__ == "__main__":
    raise SystemExit(main())
