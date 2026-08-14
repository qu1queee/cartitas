#!/usr/bin/env python3
"""Ensure card sets in draft/ and queue/ exist in en, es, and de."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
STAGES = ("draft", "queue")


def load_languages(path: Path) -> list[str]:
    with path.open(encoding="utf-8") as f:
        cfg = yaml.safe_load(f)
    return list(cfg.get("languages", {}).keys())


def find_lang_groups(base: Path) -> dict[str, set[str]]:
    groups: dict[str, set[str]] = {}
    if not base.is_dir():
        return groups
    for lang_dir in base.iterdir():
        if not lang_dir.is_dir():
            continue
        lang = lang_dir.name
        for md in lang_dir.rglob("*.md"):
            rel = md.relative_to(base)
            key = str(Path(*rel.parts[1:]))
            groups.setdefault(key, set()).add(lang)
    return groups


def validate_stage(stage: str, required_langs: list[str]) -> list[str]:
    errors: list[str] = []
    groups = find_lang_groups(ROOT / stage)
    for key, present in sorted(groups.items()):
        missing = [lang for lang in required_langs if lang not in present]
        if missing:
            have = ", ".join(sorted(present))
            need = ", ".join(missing)
            errors.append(
                f"{stage}/{key}: has [{have}], missing [{need}] — "
                f"needs all languages: {', '.join(required_langs)}"
            )
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate trilingual card sets.")
    parser.add_argument("--languages-config", type=Path, default=ROOT / "languages.yaml")
    parser.add_argument(
        "--stage",
        action="append",
        choices=STAGES,
        help="Stage to check (default: draft and queue)",
    )
    parser.add_argument("--only", help="Filter to paths containing this string")
    args = parser.parse_args()

    required = load_languages(args.languages_config)
    stages = args.stage or list(STAGES)
    errors: list[str] = []
    for stage in stages:
        errors.extend(validate_stage(stage, required))

    if args.only:
        errors = [e for e in errors if args.only in e]

    if errors:
        for err in errors:
            print(err, file=sys.stderr)
        print(
            f"\nFix: add matching files under {{stage}}/{{en,es,de}}/ for each incomplete set.",
            file=sys.stderr,
        )
        return 1

    checked = ", ".join(stages)
    print(f"All {checked} sets complete for: {', '.join(required)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
